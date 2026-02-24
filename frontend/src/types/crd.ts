export type FieldType = "string" | "number" | "boolean";

export interface FieldDefinition {
  path: string;
  label?: string;
  value?: string;
  description: string;
  type?: FieldType;
}

export interface TemplateDefinition {
  id: string;
  title: string;
  apiVersion: string;
  kind: string;
  note: string;
  defaultFields: FieldDefinition[];
  optionalFields: FieldDefinition[];
}

export interface ParseCrdResponse {
  template: TemplateDefinition;
}

export interface GenerateYamlRequest {
  apiVersion: string;
  kind: string;
  fields: FieldDefinition[];
}
