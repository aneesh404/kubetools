import type { FieldDefinition } from "../types/crd";

function parsePath(path: string): Array<string | number> {
  const segments: Array<string | number> = [];
  const parts = path.split(".");
  const pattern = /([^\[]+)|(\[(\d+)\])/g;

  parts.forEach((part) => {
    const matches = part.matchAll(pattern);
    for (const match of matches) {
      if (match[1]) {
        segments.push(match[1]);
      } else if (match[3]) {
        segments.push(Number(match[3]));
      }
    }
  });

  return segments;
}

function setDeep(target: Record<string, unknown>, path: string, rawValue: string, type?: string): void {
  const segments = parsePath(path);
  if (segments.length === 0) {
    return;
  }

  let current: unknown = target;
  for (let index = 0; index < segments.length; index += 1) {
    const key = segments[index];
    const nextKey = segments[index + 1];
    const isLast = index === segments.length - 1;

    if (isLast) {
      const resolved = parseValue(rawValue, type);
      if (typeof key === "number" && Array.isArray(current)) {
        current[key] = resolved;
      } else if (typeof key === "string" && isObject(current)) {
        current[key] = resolved;
      }
      return;
    }

    if (typeof key === "number") {
      if (!Array.isArray(current)) {
        return;
      }
      if (current[key] === undefined) {
        current[key] = typeof nextKey === "number" ? [] : {};
      }
      current = current[key];
      continue;
    }

    if (!isObject(current)) {
      return;
    }

    if (current[key] === undefined) {
      current[key] = typeof nextKey === "number" ? [] : {};
    }
    current = current[key];
  }
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function parseValue(value: string, type?: string): string | number | boolean {
  const trimmed = value.trim();

  if (type === "number" || /^-?\d+(\.\d+)?$/.test(trimmed)) {
    return Number(trimmed);
  }

  if (type === "boolean" || trimmed === "true" || trimmed === "false") {
    return trimmed === "true";
  }

  return value;
}

function formatScalar(value: unknown): string {
  if (typeof value === "string") {
    if (value === "") {
      return "\"\"";
    }
    if (/^[a-zA-Z0-9_.\-/:]+$/.test(value)) {
      return value;
    }
    return JSON.stringify(value);
  }

  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }

  return "null";
}

function toYaml(value: unknown, level = 0): string {
  const indent = "  ".repeat(level);

  if (Array.isArray(value)) {
    if (value.length === 0) {
      return `${indent}[]`;
    }

    return value
      .map((entry) => {
        if (isObject(entry) || Array.isArray(entry)) {
          return `${indent}-\n${toYaml(entry, level + 1)}`;
        }
        return `${indent}- ${formatScalar(entry)}`;
      })
      .join("\n");
  }

  if (isObject(value)) {
    const entries = Object.entries(value);
    if (entries.length === 0) {
      return `${indent}{}`;
    }

    return entries
      .map(([key, nested]) => {
        if (isObject(nested) || Array.isArray(nested)) {
          return `${indent}${key}:\n${toYaml(nested, level + 1)}`;
        }
        return `${indent}${key}: ${formatScalar(nested)}`;
      })
      .join("\n");
  }

  return `${indent}${formatScalar(value)}`;
}

export function generateYaml(
  apiVersion: string,
  kind: string,
  fields: FieldDefinition[]
): string {
  const resource: Record<string, unknown> = {
    apiVersion,
    kind
  };

  fields.forEach((field) => {
    setDeep(resource, field.path, field.value ?? "", field.type);
  });

  return `${toYaml(resource)}\n`;
}

function escapeHtml(input: string): string {
  return input
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");
}

export function highlightYaml(yaml: string): string {
  return yaml
    .split("\n")
    .map((line) => {
      const escaped = escapeHtml(line);
      const match = escaped.match(/^(\s*-?\s*)([A-Za-z0-9_.\-]+):(.*)$/);
      if (!match) {
        return escaped;
      }

      const [, prefix, key, value] = match;
      const trimmed = value.trim();
      if (!trimmed) {
        return `${prefix}<span class=\"token-key\">${key}</span>:`;
      }

      let tokenClass = "token-value";
      if (/^-?\d+(\.\d+)?$/.test(trimmed)) {
        tokenClass = "token-number";
      } else if (trimmed === "true" || trimmed === "false") {
        tokenClass = "token-boolean";
      }

      return `${prefix}<span class=\"token-key\">${key}</span>: <span class=\"${tokenClass}\">${trimmed}</span>`;
    })
    .join("\n");
}
