package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aneeshchawla/kubetools/backend/internal/config"
	"github.com/aneeshchawla/kubetools/backend/internal/models"
	"github.com/aneeshchawla/kubetools/backend/internal/services"
)

func TestSubmitCRDStoresGeneratedCustomResource(t *testing.T) {
	templateService, err := services.NewTemplateService(context.Background(), config.Config{})
	if err != nil {
		t.Logf("template service fallback: %v", err)
	}

	handler := NewCRDHandler(
		templateService,
		services.NewCRDService(),
		services.NewYAMLService(),
		&services.ManifestService{},
	)

	raw := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  group: example.io
  names:
    kind: Widget
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                size:
                  type: string
`

	payload := models.SubmitCRDRequest{
		Title: "Widget from test",
		Raw:   raw,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/crd/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SubmitCRD(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var envelope struct {
		Success bool                     `json:"success"`
		Data    models.SubmitCRDResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success {
		t.Fatalf("expected success response, got: %s", rec.Body.String())
	}

	if envelope.Data.Manifest.Kind != "Widget" {
		t.Fatalf("expected saved manifest kind Widget, got %s", envelope.Data.Manifest.Kind)
	}
	if strings.Contains(envelope.Data.Manifest.YAML, "kind: CustomResourceDefinition") {
		t.Fatalf("expected generated custom resource YAML, got CRD YAML: %s", envelope.Data.Manifest.YAML)
	}
	if !strings.Contains(envelope.Data.Manifest.YAML, "kind: Widget") {
		t.Fatalf("expected generated custom resource YAML to include kind Widget: %s", envelope.Data.Manifest.YAML)
	}
}
