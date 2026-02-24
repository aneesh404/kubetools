package services

import "testing"

func TestParseCRD(t *testing.T) {
	service := NewCRDService()
	raw := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  group: example.io
  names:
    kind: Widget
  versions:
    - name: v1
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                size:
                  type: string
                replicas:
                  type: integer
                settings:
                  type: object
                  properties:
                    region:
                      type: string
`

	result, err := service.ParseCRD(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Kind != "Widget" {
		t.Fatalf("expected kind Widget, got %s", result.Kind)
	}
	if result.APIVersion != "example.io/v1" {
		t.Fatalf("expected apiVersion example.io/v1, got %s", result.APIVersion)
	}

	paths := map[string]bool{}
	for _, field := range result.DefaultFields {
		paths[field.Path] = true
	}
	if !paths["spec.size"] {
		t.Fatalf("expected inferred field spec.size")
	}
	if !paths["spec.replicas"] {
		t.Fatalf("expected inferred field spec.replicas")
	}
	if !paths["spec.settings.region"] {
		t.Fatalf("expected inferred nested field spec.settings.region")
	}
}

func TestParseResourceYAML(t *testing.T) {
	service := NewCRDService()
	raw := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: apps
spec:
  replicas: 2
  revisionHistoryLimit: 10
`

	result, err := service.ParseCRD(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Kind != "Deployment" {
		t.Fatalf("expected kind Deployment, got %s", result.Kind)
	}
	if result.APIVersion != "apps/v1" {
		t.Fatalf("expected apiVersion apps/v1, got %s", result.APIVersion)
	}
}

func TestParseCRD_PrioritizesSignalFields(t *testing.T) {
	service := NewCRDService()
	raw := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  group: karnot.xyz
  names:
    kind: Madara
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required:
                - instanceRef
                - versionRef
              properties:
                applicationConfig:
                  type: object
                  properties:
                    coreContractAddress:
                      type: string
                    l3Enabled:
                      type: boolean
                      default: false
                env:
                  type: string
                  enum: [dev, prod]
                instanceRef:
                  type: object
                  required: [name]
                  properties:
                    name:
                      type: string
                versionRef:
                  type: object
                  required: [name]
                  properties:
                    name:
                      type: string
                bootstrapper:
                  type: object
                  properties:
                    spec:
                      type: object
                      x-kubernetes-preserve-unknown-fields: true
                      properties:
                        affinity:
                          type: object
                          properties:
                            nodeAffinity:
                              type: object
`

	result, err := service.ParseCRD(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	defaultPaths := map[string]bool{}
	for _, field := range result.DefaultFields {
		defaultPaths[field.Path] = true
	}

	if !defaultPaths["spec.instanceRef.name"] {
		t.Fatalf("expected required field spec.instanceRef.name in default form")
	}
	if !defaultPaths["spec.versionRef.name"] {
		t.Fatalf("expected required field spec.versionRef.name in default form")
	}
	if !defaultPaths["spec.applicationConfig.l3Enabled"] {
		t.Fatalf("expected default-backed field spec.applicationConfig.l3Enabled in default form")
	}
	if defaultPaths["spec.bootstrapper.spec.affinity"] {
		t.Fatalf("expected deep noisy field to be filtered out of default form")
	}
}

func TestParseCRD_IncludesServiceSeedsForAllComponents(t *testing.T) {
	service := NewCRDService()
	raw := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  group: karnot.xyz
  names:
    kind: Madara
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                bootstrapper:
                  type: object
                  properties:
                    configMaps:
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            type: string
                dna:
                  type: object
                  properties:
                    configMaps:
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            type: string
                madara:
                  type: object
                  properties:
                    configMaps:
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            type: string
                madaraFullNode:
                  type: object
                  properties:
                    configMaps:
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            type: string
                madaraOrchestrator:
                  type: object
                  properties:
                    configMaps:
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            type: string
`

	result, err := service.ParseCRD(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	paths := map[string]bool{}
	for _, field := range result.DefaultFields {
		paths[field.Path] = true
	}

	expected := []string{
		"spec.bootstrapper.configMaps[0].name",
		"spec.dna.configMaps[0].name",
		"spec.madara.configMaps[0].name",
		"spec.madaraFullNode.configMaps[0].name",
		"spec.madaraOrchestrator.configMaps[0].name",
	}

	for _, path := range expected {
		if !paths[path] {
			t.Fatalf("expected service seed field %s in default fields", path)
		}
	}
}
