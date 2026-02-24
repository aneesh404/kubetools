import type { FieldDefinition, TemplateDefinition } from "../types/crd";

const groupRegex = /^\s*group:\s*([A-Za-z0-9.-]+)\s*$/m;
const versionRegex = /^\s*version:\s*(v[0-9A-Za-z.-]+)\s*$/m;
const versionListRegex = /^\s*-\s*name:\s*(v[0-9A-Za-z.-]+)\s*$/m;
const namesKindRegex = /names:\s*(?:\n[^\n]*){0,20}\n\s*kind:\s*([A-Za-z0-9]+)\s*/m;
const kindRegex = /^\s*kind:\s*([A-Za-z0-9]+)\s*$/m;
const apiVersionRegex = /^\s*apiVersion:\s*([A-Za-z0-9./-]+)\s*$/m;
const fieldRegex = /^\s{8,}([A-Za-z][A-Za-z0-9_-]*):\s*$/gm;

export function parseCrdLocally(rawInput: string): TemplateDefinition {
  const raw = rawInput.trim();
  if (!raw) {
    throw new Error("CRD payload is empty.");
  }

  const kind = matchFirst(namesKindRegex, raw) ?? matchFirst(kindRegex, raw) ?? "CustomResource";
  const group = matchFirst(groupRegex, raw);
  const version = matchFirst(versionListRegex, raw) ?? matchFirst(versionRegex, raw);
  const apiVersion = group && version ? `${group}/${version}` : matchFirst(apiVersionRegex, raw) ?? "example.io/v1";

  const fields: FieldDefinition[] = [];
  for (const match of raw.matchAll(fieldRegex)) {
    const name = match[1];
    if (isIgnoredField(name)) {
      continue;
    }
    const path = `spec.${name}`;
    if (fields.some((field) => field.path === path)) {
      continue;
    }
    fields.push({
      path,
      description: `Inferred from parsed field '${name}'.`
    });
    if (fields.length >= 8) {
      break;
    }
  }

  if (fields.length === 0) {
    fields.push({
      path: "spec.example",
      description: "No schema fields inferred from input. Replace this with real fields."
    });
  }

  return {
    id: normalizeId(`parsed-${kind}`),
    title: `${kind} (Parsed)`,
    apiVersion,
    kind,
    note: "Parsed locally in browser. For richer schema extraction, keep backend running and provide full CRD YAML.",
    defaultFields: [
      {
        path: "metadata.name",
        value: `${kind.toLowerCase()}-sample`,
        description: "Name for this resource."
      },
      {
        path: "metadata.namespace",
        value: "default",
        description: "Namespace for this resource."
      },
      ...fields
    ],
    optionalFields: [
      {
        path: "metadata.labels.app",
        description: "Optional labels."
      },
      {
        path: "metadata.annotations.owner",
        description: "Optional annotations."
      }
    ]
  };
}

function matchFirst(regex: RegExp, input: string): string | undefined {
  const match = input.match(regex);
  if (!match || !match[1]) {
    return undefined;
  }
  return match[1].trim();
}

function isIgnoredField(field: string): boolean {
  return ["type", "properties", "items", "description", "required", "metadata", "spec", "status"].includes(field);
}

function normalizeId(input: string): string {
  const normalized = input
    .toLowerCase()
    .replaceAll("_", "-")
    .replace(/[^a-z0-9-]/g, "")
    .replace(/^-+|-+$/g, "");
  return normalized || "parsed-custom-resource";
}

