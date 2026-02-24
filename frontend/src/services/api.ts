import type {
  FieldDefinition,
  ParseCrdResponse,
  TemplateDefinition
} from "../types/crd";

export interface ManifestHistoryItem {
  id: string;
  title: string;
  resource: string;
  apiVersion: string;
  kind: string;
  yaml: string;
  createdAt: string;
  updatedAt: string;
}

export interface ValidateCrdResponse {
  valid: boolean;
  errors: string[];
  warnings: string[];
  kind?: string;
  apiVersion?: string;
}

export interface SubmitCrdResponse {
  template: TemplateDefinition;
  manifest: ManifestHistoryItem;
  validation: ValidateCrdResponse;
}

export interface ImportCrdUrlResponse {
  sourceUrl: string;
  raw: string;
  validation: ValidateCrdResponse;
}

interface ApiError {
  code: string;
  message: string;
}

interface ApiEnvelope<T> {
  success: boolean;
  data: T;
  error?: ApiError;
  timestamp: string;
}

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      headers: {
        "Content-Type": "application/json"
      },
      ...options
    });
  } catch {
    throw new Error(`Network/CORS error while calling ${path}.`);
  }

  let payload: ApiEnvelope<T> | null = null;
  try {
    payload = (await response.json()) as ApiEnvelope<T>;
  } catch {
    payload = null;
  }

  if (!response.ok) {
    throw new Error(payload?.error?.message ?? `Request failed (${response.status})`);
  }

  if (!payload || !payload.success) {
    throw new Error(payload?.error?.message ?? "API call failed");
  }

  return payload.data;
}

export async function getTemplates(): Promise<TemplateDefinition[]> {
  return request<TemplateDefinition[]>("/crd/templates");
}

export async function parseCrd(raw: string): Promise<ParseCrdResponse> {
  return request<ParseCrdResponse>("/crd/parse", {
    method: "POST",
    body: JSON.stringify({ raw })
  });
}

export async function generateYamlViaApi(args: {
  apiVersion: string;
  kind: string;
  fields: FieldDefinition[];
}): Promise<{ yaml: string }> {
  return request<{ yaml: string }>("/crd/generate-yaml", {
    method: "POST",
    body: JSON.stringify(args)
  });
}

export async function validateCrd(raw: string): Promise<ValidateCrdResponse> {
  return request<ValidateCrdResponse>("/crd/validate", {
    method: "POST",
    body: JSON.stringify({ raw })
  });
}

export async function submitCrd(args: { title: string; raw: string }): Promise<SubmitCrdResponse> {
  return request<SubmitCrdResponse>("/crd/submit", {
    method: "POST",
    body: JSON.stringify(args)
  });
}

export async function importCrdFromUrl(url: string): Promise<ImportCrdUrlResponse> {
  return request<ImportCrdUrlResponse>("/crd/import-url", {
    method: "POST",
    body: JSON.stringify({ url })
  });
}

export async function saveManifest(args: {
  title: string;
  resource: string;
  apiVersion: string;
  kind: string;
  yaml: string;
}): Promise<ManifestHistoryItem> {
  return request<ManifestHistoryItem>("/manifests", {
    method: "POST",
    body: JSON.stringify(args)
  });
}

export async function listManifests(query?: string): Promise<ManifestHistoryItem[]> {
  const params = new URLSearchParams();
  if (query && query.trim() !== "") {
    params.set("query", query.trim());
  }
  params.set("limit", "50");
  const queryString = params.toString();
  const suffix = queryString ? `?${queryString}` : "";
  return request<ManifestHistoryItem[]>(`/manifests${suffix}`);
}
