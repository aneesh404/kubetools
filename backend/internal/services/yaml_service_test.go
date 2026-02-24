package services

import (
	"strings"
	"testing"

	"github.com/aneeshchawla/kubetools/backend/internal/models"
)

func TestGenerateYAML(t *testing.T) {
	service := NewYAMLService()
	fields := []models.FieldDefinition{
		{Path: "metadata.name", Value: "demo"},
		{Path: "spec.replicas", Value: "3", Type: "number"},
		{Path: "spec.template.spec.containers[0].name", Value: "app"},
		{Path: "spec.template.spec.containers[0].image", Value: "nginx:1.27"},
	}

	output, err := service.GenerateYAML("apps/v1", "Deployment", fields)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	checks := []string{"apiVersion: apps/v1", "kind: Deployment", "name: demo", "replicas: 3"}
	for _, item := range checks {
		if !strings.Contains(output, item) {
			t.Fatalf("expected output to contain %q, got %s", item, output)
		}
	}
}
