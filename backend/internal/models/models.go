package models

import "time"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success   bool      `json:"success"`
	Data      any       `json:"data,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type FieldDefinition struct {
	Path        string `json:"path"`
	Label       string `json:"label,omitempty"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description"`
	Type        string `json:"type,omitempty"`
}

type TemplateDefinition struct {
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	APIVersion     string            `json:"apiVersion"`
	Kind           string            `json:"kind"`
	Note           string            `json:"note"`
	DefaultFields  []FieldDefinition `json:"defaultFields"`
	OptionalFields []FieldDefinition `json:"optionalFields"`
}

type ParseCRDRequest struct {
	Raw string `json:"raw"`
}

type ParseCRDResponse struct {
	Template TemplateDefinition `json:"template"`
}

type ValidateCRDRequest struct {
	Raw string `json:"raw"`
}

type ValidateCRDResponse struct {
	Valid      bool     `json:"valid"`
	Errors     []string `json:"errors"`
	Warnings   []string `json:"warnings"`
	Kind       string   `json:"kind,omitempty"`
	APIVersion string   `json:"apiVersion,omitempty"`
}

type SubmitCRDRequest struct {
	Title string `json:"title"`
	Raw   string `json:"raw"`
}

type SubmitCRDResponse struct {
	Template   TemplateDefinition  `json:"template"`
	Manifest   ManifestRecord      `json:"manifest"`
	Validation ValidateCRDResponse `json:"validation"`
}

type ImportCRDURLRequest struct {
	URL string `json:"url"`
}

type ImportCRDURLResponse struct {
	SourceURL  string              `json:"sourceUrl"`
	Raw        string              `json:"raw"`
	Validation ValidateCRDResponse `json:"validation"`
}

type GenerateYAMLRequest struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Fields     []FieldDefinition `json:"fields"`
}

type GenerateYAMLResponse struct {
	YAML string `json:"yaml"`
}

type SaveManifestRequest struct {
	Title      string `json:"title"`
	Resource   string `json:"resource"`
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	YAML       string `json:"yaml"`
}

type ManifestRecord struct {
	ID         string    `json:"id" bson:"_id"`
	Title      string    `json:"title" bson:"title"`
	Resource   string    `json:"resource" bson:"resource"`
	APIVersion string    `json:"apiVersion" bson:"apiVersion"`
	Kind       string    `json:"kind" bson:"kind"`
	YAML       string    `json:"yaml" bson:"yaml"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" bson:"updatedAt"`
}
