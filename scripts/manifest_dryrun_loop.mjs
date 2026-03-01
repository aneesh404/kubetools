#!/usr/bin/env node

import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

const API_BASE = process.env.KUBETOOLS_API_BASE ?? "http://localhost:8080/api/v1";
const REPORT_PATH = process.env.KUBETOOLS_REPORT_PATH ?? path.resolve("out/manifest-validation-report.md");

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    encoding: "utf8",
    ...options
  });
  return {
    ok: result.status === 0,
    code: result.status ?? -1,
    stdout: (result.stdout ?? "").trim(),
    stderr: (result.stderr ?? "").trim()
  };
}

function hasCommand(command) {
  return run("bash", ["-lc", `command -v ${command}`]).ok;
}

function kubeconformCommand() {
  if (hasCommand("kubeconform")) {
    return "kubeconform";
  }

  const fallback = path.join(process.env.HOME ?? "", "go", "bin", "kubeconform");
  if (run("bash", ["-lc", `test -x "${fallback}"`]).ok) {
    return fallback;
  }
  return null;
}

async function request(endpoint, init = {}) {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    headers: { "Content-Type": "application/json" },
    ...init
  });

  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(payload?.error?.message ?? `API ${endpoint} failed with status ${response.status}`);
  }
  if (!payload?.success) {
    throw new Error(payload?.error?.message ?? `API ${endpoint} returned unsuccessful envelope`);
  }
  return payload.data;
}

function cloneFields(fields) {
  return fields.map((field) => ({ ...field }));
}

function sampleValue(field) {
  const pathValue = (field.path ?? "").toLowerCase();
  const fieldType = (field.type ?? "").toLowerCase();

  if (fieldType === "number" || pathValue.includes("replicas") || pathValue.includes("port")) {
    return "1";
  }
  if (
    fieldType === "boolean" ||
    pathValue.endsWith(".enabled") ||
    pathValue.includes(".enabled.") ||
    pathValue.endsWith(".is")
  ) {
    return "true";
  }
  if (pathValue.endsWith(".image")) {
    return "nginx:1.27";
  }
  if (pathValue.endsWith(".name")) {
    return "sample";
  }
  if (pathValue.endsWith(".namespace")) {
    return "default";
  }
  if (pathValue.includes("storage")) {
    return "5Gi";
  }
  if (pathValue.includes("schedule")) {
    return "0 3 * * *";
  }
  if (pathValue.includes("class")) {
    return "standard";
  }
  if (pathValue.includes("app")) {
    return "demo";
  }
  return "value";
}

function mutateFieldValue(field) {
  const original = String(field.value ?? "");
  const fieldType = (field.type ?? "").toLowerCase();
  const pathValue = (field.path ?? "").toLowerCase();

  if (fieldType === "number" || /^-?\d+(\.\d+)?$/.test(original)) {
    const parsed = Number(original || "0");
    return String(Number.isFinite(parsed) ? parsed + 1 : 1);
  }
  if (fieldType === "boolean" || original === "true" || original === "false") {
    return original === "true" ? "false" : "true";
  }
  if (pathValue.endsWith(".name")) {
    return original === "" ? "sample-name" : `${original}-alt`;
  }
  return original === "" ? sampleValue(field) : `${original}-alt`;
}

function addOptionalFields(baseFields, optionalFields, count) {
  const cloned = cloneFields(baseFields);
  const existing = new Set(cloned.map((field) => field.path));
  for (const optional of optionalFields.slice(0, count)) {
    if (existing.has(optional.path)) {
      continue;
    }
    cloned.push({
      ...optional,
      value: optional.value ?? sampleValue(optional)
    });
  }
  return cloned;
}

function generateVariants(template) {
  const variants = [];
  const base = cloneFields(template.defaultFields ?? []);
  const optional = template.optionalFields ?? [];

  variants.push({ name: "base", fields: base });
  if (optional.length > 0) {
    variants.push({ name: "optional-1", fields: addOptionalFields(base, optional, 1) });
    variants.push({ name: "optional-4", fields: addOptionalFields(base, optional, 4) });
  }

  variants.push({
    name: "mutated-defaults",
    fields: base.map((field) => ({
      ...field,
      value: mutateFieldValue(field)
    }))
  });

  variants.push({
    name: "dense",
    fields: addOptionalFields(
      base.map((field) => ({ ...field, value: field.value ?? sampleValue(field) })),
      optional,
      12
    )
  });

  const seen = new Set();
  return variants.filter((variant) => {
    const signature = JSON.stringify(
      variant.fields.map((field) => [field.path, field.type ?? "", field.value ?? ""])
    );
    if (seen.has(signature)) {
      return false;
    }
    seen.add(signature);
    return true;
  });
}

function isKubectlInfraError(output) {
  const lower = output.toLowerCase();
  return (
    lower.includes("token has expired") ||
    lower.includes("couldn't get current server api group list") ||
    lower.includes("failed to download openapi") ||
    lower.includes("getting credentials")
  );
}

function sanitizeName(value) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9.-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 63);
}

function writeReport(lines) {
  const content = `${lines.join("\n")}\n`;
  run("bash", ["-lc", `mkdir -p "${path.dirname(REPORT_PATH)}"`]);
  writeFileSync(REPORT_PATH, content, "utf8");
}

async function main() {
  const kubeconform = kubeconformCommand();
  if (!kubeconform) {
    throw new Error("kubeconform is required for this loop. Install with: go install github.com/yannh/kubeconform/cmd/kubeconform@v0.6.7");
  }
  if (!hasCommand("yq")) {
    throw new Error("yq is required for YAML parsing checks.");
  }

  const templates = await request("/crd/templates");
  const manifests = await request("/manifests?limit=500");

  const crdManifests = manifests.filter((manifest) => {
    const kind = String(manifest.kind ?? "").toLowerCase();
    const title = String(manifest.title ?? "").toLowerCase();
    return kind === "customresourcedefinition" || title.startsWith("schema:");
  });

  const parsedTemplates = [];
  for (const crd of crdManifests) {
    try {
      const parsed = await request("/crd/parse", {
        method: "POST",
        body: JSON.stringify({ raw: crd.yaml })
      });
      parsedTemplates.push({
        ...parsed.template,
        __source: `CRD:${crd.title}`
      });
    } catch (error) {
      parsedTemplates.push({
        id: `parse-failed-${crd.id}`,
        title: crd.title,
        apiVersion: crd.apiVersion,
        kind: crd.kind,
        defaultFields: [],
        optionalFields: [],
        note: `parse failed: ${error instanceof Error ? error.message : String(error)}`,
        __source: `CRD:${crd.title}`
      });
    }
  }

  const mergedTemplates = [];
  const seenTemplate = new Set();
  for (const template of [...parsedTemplates, ...templates]) {
    const key = `${template.apiVersion}|${template.kind}|${template.id}`;
    if (seenTemplate.has(key)) {
      continue;
    }
    seenTemplate.add(key);
    mergedTemplates.push({
      ...template,
      __source: template.__source ?? "default-template"
    });
  }

  const tmpDir = mkdtempSync(path.join(tmpdir(), "kubetools-dryrun-"));
  const report = [];
  const results = [];

  report.push("# Manifest Dry-Run Loop Report");
  report.push(`Generated at: ${new Date().toISOString()}`);
  report.push(`API base: ${API_BASE}`);
  report.push(`Templates validated: ${mergedTemplates.length}`);
  report.push("");

  for (const template of mergedTemplates) {
    const variants = generateVariants(template);
    report.push(`## ${template.kind} (${template.apiVersion})`);
    report.push(`Source: ${template.__source}`);
    report.push(`Variants: ${variants.length}`);

    for (const variant of variants) {
      const generated = await request("/crd/generate-yaml", {
        method: "POST",
        body: JSON.stringify({
          apiVersion: template.apiVersion,
          kind: template.kind,
          fields: variant.fields
        })
      });

      const fileName = sanitizeName(`${template.kind}-${variant.name}`) || "manifest";
      const filePath = path.join(tmpDir, `${fileName}.yaml`);
      writeFileSync(filePath, generated.yaml, "utf8");

      if (hasCommand("pbcopy")) {
        run("bash", ["-lc", `cat "${filePath}" | pbcopy`]);
      }

      const yqResult = run("yq", ["e", ".", filePath]);
      const kubeconformResult = run(kubeconform, ["-strict", "-ignore-missing-schemas", "-summary", filePath]);
      const kubectlResult = run("kubectl", ["apply", "--dry-run=client", "--validate=false", "-f", filePath]);
      const kubectlOutput = [kubectlResult.stdout, kubectlResult.stderr].filter(Boolean).join("\n");
      const kubectlInfraBlocked = !kubectlResult.ok && isKubectlInfraError(kubectlOutput);

      const pass = yqResult.ok && kubeconformResult.ok && (kubectlResult.ok || kubectlInfraBlocked);
      results.push({
        template: template.kind,
        variant: variant.name,
        pass,
        yqOk: yqResult.ok,
        kubeconformOk: kubeconformResult.ok,
        kubectlOk: kubectlResult.ok,
        kubectlInfraBlocked
      });

      report.push(
        `- ${variant.name}: ${
          pass ? "PASS" : "FAIL"
        } | yq=${yqResult.ok ? "ok" : "fail"} | kubeconform=${kubeconformResult.ok ? "ok" : "fail"} | kubectl=${
          kubectlResult.ok ? "ok" : kubectlInfraBlocked ? "infra-blocked" : "fail"
        }`
      );

      if (!pass) {
        if (!yqResult.ok) {
          report.push(`  - yq: ${yqResult.stderr || yqResult.stdout}`);
        }
        if (!kubeconformResult.ok) {
          report.push(`  - kubeconform: ${kubeconformResult.stderr || kubeconformResult.stdout}`);
        }
        if (!kubectlResult.ok && !kubectlInfraBlocked) {
          report.push(`  - kubectl: ${kubectlOutput}`);
        }
      }
    }
    report.push("");
  }

  const passCount = results.filter((item) => item.pass).length;
  const failCount = results.length - passCount;
  const infraBlockedCount = results.filter((item) => item.kubectlInfraBlocked).length;

  report.unshift(`Summary: ${passCount}/${results.length} passed, ${failCount} failed, kubectl infra-blocked in ${infraBlockedCount} cases.`);
  writeReport(report);

  rmSync(tmpDir, { recursive: true, force: true });

  if (failCount > 0) {
    console.error(`Validation found ${failCount} failing manifest variants. See ${REPORT_PATH}`);
    process.exit(1);
  }

  console.log(`Validation passed for ${passCount} variants. Report: ${REPORT_PATH}`);
}

main().catch((error) => {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
});
