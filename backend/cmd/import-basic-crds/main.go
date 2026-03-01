package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aneeshchawla/kubetools/backend/internal/config"
	"github.com/aneeshchawla/kubetools/backend/internal/services"
	"gopkg.in/yaml.v3"
)

var (
	slugRegex = regexp.MustCompile(`[^a-z0-9]+`)
	sources   = []string{
		"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/application-crd.yaml",
		"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/appproject-crd.yaml",
		"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/applicationset-crd.yaml",
		"https://raw.githubusercontent.com/external-secrets/external-secrets/main/config/crds/bases/external-secrets.io_externalsecrets.yaml",
		"https://raw.githubusercontent.com/external-secrets/external-secrets/main/config/crds/bases/external-secrets.io_secretstores.yaml",
		"https://raw.githubusercontent.com/external-secrets/external-secrets/main/config/crds/bases/external-secrets.io_clustersecretstores.yaml",
		"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml",
		"https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml",
		"https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.crds.yaml",
	}
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	templateService, err := services.NewTemplateService(ctx, cfg)
	if err != nil {
		panic(fmt.Errorf("template service requires MongoDB connection: %w", err))
	}
	defer templateService.Close(context.Background())

	crdService := services.NewCRDService()
	client := &http.Client{Timeout: 20 * time.Second}

	imported := make([]string, 0, 32)
	for _, source := range sources {
		raw, err := fetchSource(ctx, client, source)
		if err != nil {
			fmt.Printf("[WARN] fetch failed: %s (%v)\n", source, err)
			continue
		}

		crdDocs, err := extractCRDDocuments(raw)
		if err != nil {
			fmt.Printf("[WARN] decode failed: %s (%v)\n", source, err)
			continue
		}
		if len(crdDocs) == 0 {
			fmt.Printf("[WARN] no CRDs found in: %s\n", source)
			continue
		}

		for _, doc := range crdDocs {
			template, err := crdService.ParseCRD(doc)
			if err != nil {
				fmt.Printf("[WARN] parse failed from %s: %v\n", source, err)
				continue
			}

			group := strings.SplitN(template.APIVersion, "/", 2)[0]
			template.ID = normalizeTemplateID(template.Kind, group)
			template.Title = fmt.Sprintf("%s (%s)", template.Kind, group)
			template.Note = "Imported from official upstream CRD source."

			if err := templateService.Upsert(ctx, template); err != nil {
				fmt.Printf("[WARN] upsert failed for %s from %s: %v\n", template.ID, source, err)
				continue
			}

			imported = append(imported, template.ID)
		}
	}

	if len(imported) == 0 {
		fmt.Println("No templates were imported.")
		return
	}

	unique := uniqueStrings(imported)
	fmt.Printf("Imported/updated %d templates in MongoDB.\n", len(unique))
	for _, id := range unique {
		fmt.Printf("- %s\n", id)
	}
}

func fetchSource(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "kubetools-basic-crd-importer/1.0")
	req.Header.Set("Accept", "application/yaml, text/plain, */*")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func extractCRDDocuments(raw string) ([]string, error) {
	decoder := yaml.NewDecoder(strings.NewReader(raw))
	out := make([]string, 0, 8)
	for {
		doc := map[string]any{}
		err := decoder.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(doc) == 0 {
			continue
		}
		if asString(doc["kind"]) != "CustomResourceDefinition" {
			continue
		}
		encoded, err := yaml.Marshal(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, string(encoded))
	}
	return out, nil
}

func asString(value any) string {
	v, _ := value.(string)
	return strings.TrimSpace(v)
}

func normalizeTemplateID(kind string, group string) string {
	raw := strings.ToLower(strings.TrimSpace(kind + "-" + group))
	raw = slugRegex.ReplaceAllString(raw, "-")
	raw = strings.Trim(raw, "-")
	if raw == "" {
		return "parsed-imported-crd"
	}
	return "parsed-" + raw
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
