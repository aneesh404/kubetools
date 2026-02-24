package services

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aneeshchawla/kubetools/backend/internal/models"
	"gopkg.in/yaml.v3"
)

var (
	regexGroup      = regexp.MustCompile(`(?m)^\s*group:\s*([A-Za-z0-9.-]+)\s*$`)
	regexVersion    = regexp.MustCompile(`(?m)^\s*version:\s*(v[0-9A-Za-z.-]+)\s*$`)
	regexVersionRef = regexp.MustCompile(`(?m)^\s*-\s*name:\s*(v[0-9A-Za-z.-]+)\s*$`)
	regexKind       = regexp.MustCompile(`(?s)names:\s*(?:\n[^\n]*){0,20}\n\s*kind:\s*([A-Za-z0-9]+)\s*`)
	regexField      = regexp.MustCompile(`(?m)^\s{8,}([A-Za-z][A-Za-z0-9_-]*):\s*$`)
)

type CRDService struct{}

func NewCRDService() *CRDService {
	return &CRDService{}
}

func (s *CRDService) ParseCRD(raw string) (models.TemplateDefinition, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return models.TemplateDefinition{}, errors.New("CRD payload is empty")
	}

	if structured, ok := parseStructuredYAML(raw); ok {
		return structured, nil
	}

	return parseWithRegexFallback(raw), nil
}

func (s *CRDService) ValidateCRD(raw string) models.ValidateCRDResponse {
	result := models.ValidateCRDResponse{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		result.Errors = append(result.Errors, "CRD payload is empty.")
		return result
	}

	docs, err := decodeYAMLDocuments(raw)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("YAML parse error: %v", err))
		return result
	}
	root, ok := selectPrimaryResourceDoc(docs)
	if !ok || len(root) == 0 {
		result.Errors = append(result.Errors, "YAML payload has no valid resource documents.")
		return result
	}

	result.Kind = asString(root["kind"])
	result.APIVersion = asString(root["apiVersion"])

	if result.Kind == "" {
		result.Errors = append(result.Errors, "Missing required top-level field: kind")
	}
	if result.APIVersion == "" {
		result.Errors = append(result.Errors, "Missing required top-level field: apiVersion")
	}

	if strings.EqualFold(result.Kind, "CustomResourceDefinition") {
		specMap, ok := root["spec"].(map[string]any)
		if !ok {
			result.Errors = append(result.Errors, "Missing required object: spec")
		} else {
			if asString(specMap["group"]) == "" {
				result.Errors = append(result.Errors, "Missing required CRD field: spec.group")
			}

			namesMap, _ := specMap["names"].(map[string]any)
			if asString(namesMap["kind"]) == "" {
				result.Errors = append(result.Errors, "Missing required CRD field: spec.names.kind")
			}

			version := asString(specMap["version"])
			versions, _ := specMap["versions"].([]any)
			if version == "" && len(versions) == 0 {
				result.Errors = append(result.Errors, "Missing required CRD version field: spec.version or spec.versions")
			}
			if len(versions) > 0 {
				for i, item := range versions {
					versionMap, ok := item.(map[string]any)
					if !ok || asString(versionMap["name"]) == "" {
						result.Errors = append(result.Errors, fmt.Sprintf("Invalid spec.versions[%d]: missing name", i))
					}
				}
			}

			if !hasCRDSchema(root) {
				result.Warnings = append(result.Warnings, "CRD schema not found. Add openAPIV3Schema for richer field guidance.")
			}
		}
	} else if result.Kind != "" {
		result.Warnings = append(result.Warnings, "Input kind is not CustomResourceDefinition. It will still be accepted.")
	}

	result.Valid = len(result.Errors) == 0
	return result
}

func (s *CRDService) FetchCRDFromURL(rawURL string) (string, string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", "", errors.New("url is required")
	}

	parsed, err := neturl.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", "", errors.New("only http and https urls are supported")
	}
	if parsed.Hostname() == "" {
		return "", "", errors.New("url hostname is required")
	}

	normalized := normalizeSourceURL(parsed)
	client := &http.Client{Timeout: 12 * time.Second}
	req, err := http.NewRequest(http.MethodGet, normalized, nil)
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "kubebuilder-crd-import/1.0")
	req.Header.Set("Accept", "text/plain, application/yaml, application/x-yaml, */*")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", "", fmt.Errorf("fetch failed with status %d", resp.StatusCode)
	}

	const maxBytes = 2 * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return "", "", fmt.Errorf("read response: %w", err)
	}
	if len(body) > maxBytes {
		return "", "", errors.New("document is too large (max 2MB)")
	}

	contents := strings.TrimSpace(string(body))
	if contents == "" {
		return "", "", errors.New("document is empty")
	}

	return normalized, contents, nil
}

func normalizeSourceURL(parsed *neturl.URL) string {
	host := strings.ToLower(parsed.Hostname())
	if host == "github.com" {
		segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(segments) >= 5 && segments[2] == "blob" {
			owner := segments[0]
			repo := segments[1]
			branch := segments[3]
			filePath := path.Join(segments[4:]...)
			return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, filePath)
		}
	}
	return parsed.String()
}

func parseStructuredYAML(raw string) (models.TemplateDefinition, bool) {
	docs, err := decodeYAMLDocuments(raw)
	if err != nil {
		return models.TemplateDefinition{}, false
	}
	root, ok := selectPrimaryResourceDoc(docs)
	if !ok || len(root) == 0 {
		return models.TemplateDefinition{}, false
	}

	topKind := asString(root["kind"])
	if strings.EqualFold(topKind, "CustomResourceDefinition") {
		return parseCRDDocument(root), true
	}

	if topKind != "" {
		return parseArbitraryResource(root), true
	}

	return models.TemplateDefinition{}, false
}

func parseCRDDocument(root map[string]any) models.TemplateDefinition {
	kind := asString(nested(root, "spec", "names", "kind"))
	if kind == "" {
		kind = "CustomResource"
	}

	group := asString(nested(root, "spec", "group"))
	version := asString(nested(root, "spec", "version"))

	defaultFields, optionalFields, schemaVersion := extractCRDSpecFields(root)
	if version == "" {
		version = schemaVersion
	}
	if version == "" {
		version = firstVersion(nested(root, "spec", "versions"))
	}

	apiVersion := "example.io/v1"
	if group != "" && version != "" {
		apiVersion = fmt.Sprintf("%s/%s", group, version)
	}

	if len(defaultFields) == 0 {
		defaultFields = []models.FieldDefinition{
			{
				Path:        "spec.example",
				Description: "No explicit schema fields found in CRD. Replace with a valid spec field.",
			},
		}
	}
	if specSchema, _ := selectSpecSchema(root); specSchema != nil {
		serviceSeeds := extractServiceSeedFields(specSchema)
		defaultFields = dedupeFields(append(serviceSeeds, defaultFields...))
	}

	optionalFields = dedupeFields(append(optionalFields,
		models.FieldDefinition{Path: "metadata.labels.app", Description: "Optional labels for grouping and selectors."},
		models.FieldDefinition{Path: "metadata.annotations.owner", Description: "Optional metadata annotation for ownership."},
	))

	return models.TemplateDefinition{
		ID:         normalizeID("parsed-" + kind),
		Title:      kind + " (Parsed)",
		APIVersion: apiVersion,
		Kind:       kind,
		Note:       "Generated from CRD schema. Prioritizing required and high-signal fields for cleaner authoring.",
		DefaultFields: append([]models.FieldDefinition{
			{Path: "metadata.name", Value: strings.ToLower(kind) + "-sample", Description: "Name for this custom resource."},
			{Path: "metadata.namespace", Value: "default", Description: "Namespace for this custom resource."},
		}, defaultFields...),
		OptionalFields: optionalFields,
	}
}

func parseArbitraryResource(root map[string]any) models.TemplateDefinition {
	kind := asString(root["kind"])
	apiVersion := asString(root["apiVersion"])
	if apiVersion == "" {
		apiVersion = "example.io/v1"
	}

	name := asString(nested(root, "metadata", "name"))
	if name == "" {
		name = strings.ToLower(kind) + "-sample"
	}
	namespace := asString(nested(root, "metadata", "namespace"))
	if namespace == "" {
		namespace = "default"
	}

	fields := []models.FieldDefinition{
		{Path: "metadata.name", Value: name, Description: "Name for this resource."},
		{Path: "metadata.namespace", Value: namespace, Description: "Namespace for this resource."},
	}

	specMap, _ := nested(root, "spec").(map[string]any)
	if specMap != nil {
		specKeys := make([]string, 0, len(specMap))
		for key := range specMap {
			specKeys = append(specKeys, key)
		}
		sort.Strings(specKeys)

		for _, key := range specKeys {
			if len(fields) >= 10 {
				break
			}
			value := specMap[key]
			field := models.FieldDefinition{
				Path:        "spec." + key,
				Description: fmt.Sprintf("Inferred from resource spec field '%s'.", key),
			}
			switch typed := value.(type) {
			case string:
				field.Value = typed
			case bool:
				field.Value = fmt.Sprintf("%t", typed)
				field.Type = "boolean"
			case int, int32, int64, float32, float64:
				field.Value = fmt.Sprintf("%v", typed)
				field.Type = "number"
			}
			fields = append(fields, field)
		}
	}

	return models.TemplateDefinition{
		ID:             normalizeID("parsed-" + kind),
		Title:          kind + " (Parsed)",
		APIVersion:     apiVersion,
		Kind:           kind,
		Note:           "Parsed from a Kubernetes object. Fields were inferred from current manifest.",
		DefaultFields:  fields,
		OptionalFields: []models.FieldDefinition{},
	}
}

type schemaFieldCandidate struct {
	Field      models.FieldDefinition
	Required   bool
	Depth      int
	HasDefault bool
}

func extractCRDSpecFields(root map[string]any) ([]models.FieldDefinition, []models.FieldDefinition, string) {
	specSchema, schemaVersion := selectSpecSchema(root)
	if specSchema == nil {
		return nil, nil, schemaVersion
	}

	properties, _ := specSchema["properties"].(map[string]any)
	if len(properties) == 0 {
		return nil, nil, schemaVersion
	}

	requiredSet := parseRequiredSet(specSchema["required"])
	collected := make([]schemaFieldCandidate, 0, 96)
	collectSchemaFields("spec", properties, requiredSet, 0, 4, 420, &collected)

	if len(collected) == 0 {
		return nil, nil, schemaVersion
	}

	collected = dedupeCandidates(collected)
	sort.SliceStable(collected, func(i, j int) bool {
		left := collected[i]
		right := collected[j]

		if left.Required != right.Required {
			return left.Required
		}
		if left.HasDefault != right.HasDefault {
			return left.HasDefault
		}
		if left.Depth != right.Depth {
			return left.Depth < right.Depth
		}
		return left.Field.Path < right.Field.Path
	})

	defaults := make([]models.FieldDefinition, 0, 16)
	optionals := make([]models.FieldDefinition, 0, len(collected))
	seenDefault := make(map[string]struct{}, 16)

	const maxDefaultFields = 64
	for _, candidate := range collected {
		shouldDefault := candidate.Required || candidate.HasDefault
		if shouldDefault && len(defaults) < maxDefaultFields {
			defaults = append(defaults, candidate.Field)
			seenDefault[candidate.Field.Path] = struct{}{}
			continue
		}
		optionals = append(optionals, candidate.Field)
	}

	for _, candidate := range collected {
		if len(defaults) >= maxDefaultFields {
			break
		}
		if _, exists := seenDefault[candidate.Field.Path]; exists {
			continue
		}
		defaults = append(defaults, candidate.Field)
		seenDefault[candidate.Field.Path] = struct{}{}
	}

	finalOptionals := make([]models.FieldDefinition, 0, len(optionals))
	for _, field := range optionals {
		if _, exists := seenDefault[field.Path]; exists {
			continue
		}
		finalOptionals = append(finalOptionals, field)
	}

	defaults = ensureTopLevelSpecCoverage(specSchema, defaults, collected)

	return defaults, finalOptionals, schemaVersion
}

func collectSchemaFields(
	prefix string,
	properties map[string]any,
	requiredSet map[string]bool,
	depth int,
	maxDepth int,
	limit int,
	out *[]schemaFieldCandidate,
) {
	if len(*out) >= limit {
		return
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if len(*out) >= limit {
			return
		}

		node, ok := properties[key].(map[string]any)
		if !ok {
			continue
		}

		path := prefix + "." + key
		nodeType := asString(node["type"])
		description := asString(node["description"])
		if description == "" {
			description = fmt.Sprintf("Inferred from CRD schema field '%s'.", key)
		}

		if hasPreservedUnknownFields(node) && depth >= 2 {
			continue
		}
		if depth > maxDepth {
			continue
		}

		defaultValue, hasDefault := inferSchemaDefault(node)
		fieldType := normalizeFieldType(nodeType)
		isRequired := requiredSet[key]

		if nodeType == "array" {
			items, _ := node["items"].(map[string]any)
			itemProps, _ := items["properties"].(map[string]any)
			if len(itemProps) > 0 && depth <= maxDepth && !hasPreservedUnknownFields(items) {
				itemRequired := parseRequiredSet(items["required"])
				collectSchemaFields(path+"[0]", itemProps, itemRequired, depth, maxDepth, limit, out)
				continue
			}

			candidate := schemaFieldCandidate{
				Field: models.FieldDefinition{
					Path:        path + "[0]",
					Type:        fieldType,
					Value:       defaultValue,
					Description: description,
				},
				Required:   isRequired,
				Depth:      depth,
				HasDefault: hasDefault,
			}
			*out = append(*out, candidate)
			continue
		}

		nestedProps, hasNested := node["properties"].(map[string]any)
		if hasNested && len(nestedProps) > 0 {
			if depth < maxDepth && !hasPreservedUnknownFields(node) {
				nestedRequired := parseRequiredSet(node["required"])
				collectSchemaFields(path, nestedProps, nestedRequired, depth+1, maxDepth, limit, out)
			}
			continue
		}

		if nodeType == "object" || (nodeType == "" && node["additionalProperties"] != nil) {
			if field, ok := mapSeedField(path, node, description); ok {
				*out = append(*out, schemaFieldCandidate{
					Field:      field,
					Required:   isRequired,
					Depth:      depth,
					HasDefault: hasDefault,
				})
			}
			continue
		}

		*out = append(*out, schemaFieldCandidate{
			Field: models.FieldDefinition{
				Path:        path,
				Type:        fieldType,
				Value:       defaultValue,
				Description: description,
			},
			Required:   isRequired,
			Depth:      depth,
			HasDefault: hasDefault,
		})
	}
}

func selectSpecSchema(root map[string]any) (map[string]any, string) {
	specMap, _ := root["spec"].(map[string]any)
	if specMap == nil {
		return nil, ""
	}

	versions, _ := specMap["versions"].([]any)
	if len(versions) > 0 {
		selected, versionName := selectVersionSchema(versions)
		if selected != nil {
			return selected, versionName
		}
	}

	validation, _ := specMap["validation"].(map[string]any)
	openSchema, _ := validation["openAPIV3Schema"].(map[string]any)
	properties, _ := openSchema["properties"].(map[string]any)
	specSchema, _ := properties["spec"].(map[string]any)
	return specSchema, asString(specMap["version"])
}

func selectVersionSchema(versions []any) (map[string]any, string) {
	type candidate struct {
		Schema  map[string]any
		Name    string
		Storage bool
		Served  bool
	}

	picked := make([]candidate, 0, len(versions))
	for _, entry := range versions {
		versionMap, _ := entry.(map[string]any)
		if versionMap == nil {
			continue
		}

		name := asString(versionMap["name"])
		schemaMap, _ := versionMap["schema"].(map[string]any)
		openSchema, _ := schemaMap["openAPIV3Schema"].(map[string]any)
		properties, _ := openSchema["properties"].(map[string]any)
		specSchema, _ := properties["spec"].(map[string]any)
		if specSchema == nil {
			continue
		}

		candidateItem := candidate{
			Schema:  specSchema,
			Name:    name,
			Storage: asBool(versionMap["storage"]),
			Served:  asBool(versionMap["served"]),
		}
		picked = append(picked, candidateItem)
	}

	if len(picked) == 0 {
		return nil, ""
	}

	for _, item := range picked {
		if item.Storage {
			return item.Schema, item.Name
		}
	}
	for _, item := range picked {
		if item.Served {
			return item.Schema, item.Name
		}
	}

	return picked[0].Schema, picked[0].Name
}

func parseRequiredSet(value any) map[string]bool {
	required := make(map[string]bool)
	list, _ := value.([]any)
	for _, item := range list {
		name := asString(item)
		if name == "" {
			continue
		}
		required[name] = true
	}
	return required
}

func inferSchemaDefault(node map[string]any) (string, bool) {
	if node == nil {
		return "", false
	}
	if value, exists := node["default"]; exists {
		text := formatDefaultValue(value)
		return text, true
	}

	enumValues, _ := node["enum"].([]any)
	if len(enumValues) > 0 {
		text := formatDefaultValue(enumValues[0])
		return text, true
	}

	return "", false
}

func formatDefaultValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case bool:
		return fmt.Sprintf("%t", typed)
	case int:
		return fmt.Sprintf("%d", typed)
	case int32:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case float32:
		return fmt.Sprintf("%g", typed)
	case float64:
		return fmt.Sprintf("%g", typed)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func hasPreservedUnknownFields(node map[string]any) bool {
	if node == nil {
		return false
	}
	return asBool(node["x-kubernetes-preserve-unknown-fields"])
}

func dedupeCandidates(items []schemaFieldCandidate) []schemaFieldCandidate {
	merged := make(map[string]schemaFieldCandidate, len(items))
	order := make([]string, 0, len(items))
	for _, item := range items {
		existing, found := merged[item.Field.Path]
		if !found {
			merged[item.Field.Path] = item
			order = append(order, item.Field.Path)
			continue
		}

		if item.Required {
			existing.Required = true
		}
		if item.HasDefault && !existing.HasDefault {
			existing.HasDefault = true
			existing.Field.Value = item.Field.Value
		}
		if existing.Field.Description == "" {
			existing.Field.Description = item.Field.Description
		}
		if existing.Field.Type == "" {
			existing.Field.Type = item.Field.Type
		}
		if item.Depth < existing.Depth {
			existing.Depth = item.Depth
		}
		merged[item.Field.Path] = existing
	}

	out := make([]schemaFieldCandidate, 0, len(order))
	for _, key := range order {
		out = append(out, merged[key])
	}
	return out
}

func dedupeFields(fields []models.FieldDefinition) []models.FieldDefinition {
	seen := make(map[string]struct{}, len(fields))
	out := make([]models.FieldDefinition, 0, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(field.Path) == "" {
			continue
		}
		if _, exists := seen[field.Path]; exists {
			continue
		}
		seen[field.Path] = struct{}{}
		out = append(out, field)
	}
	return out
}

func mapSeedField(path string, node map[string]any, description string) (models.FieldDefinition, bool) {
	additional, ok := node["additionalProperties"].(map[string]any)
	if !ok || additional == nil {
		return models.FieldDefinition{}, false
	}
	defaultValue, _ := inferSchemaDefault(additional)
	return models.FieldDefinition{
		Path:        path + ".exampleKey",
		Type:        normalizeFieldType(asString(additional["type"])),
		Value:       defaultValue,
		Description: description + " (map entry key/value).",
	}, true
}

func ensureTopLevelSpecCoverage(
	specSchema map[string]any,
	defaults []models.FieldDefinition,
	collected []schemaFieldCandidate,
) []models.FieldDefinition {
	properties, _ := specSchema["properties"].(map[string]any)
	if len(properties) == 0 {
		return defaults
	}

	candidatesByTop := make(map[string][]models.FieldDefinition, len(properties))
	for _, candidate := range collected {
		top := topLevelSpecKey(candidate.Field.Path)
		if top == "" {
			continue
		}
		candidatesByTop[top] = append(candidatesByTop[top], candidate.Field)
	}

	existingTop := make(map[string]bool, len(defaults))
	for _, field := range defaults {
		top := topLevelSpecKey(field.Path)
		if top != "" {
			existingTop[top] = true
		}
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	extra := make([]models.FieldDefinition, 0, len(keys))
	for _, key := range keys {
		if existingTop[key] {
			continue
		}

		if candidates := candidatesByTop[key]; len(candidates) > 0 {
			extra = append(extra, candidates[0])
			continue
		}

		node, _ := properties[key].(map[string]any)
		if field, ok := buildServiceSeedField(key, node); ok {
			extra = append(extra, field)
		}
	}

	if len(extra) == 0 {
		return defaults
	}

	return dedupeFields(append(defaults, extra...))
}

func topLevelSpecKey(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) < 2 || parts[0] != "spec" {
		return ""
	}
	key := parts[1]
	if idx := strings.Index(key, "["); idx >= 0 {
		key = key[:idx]
	}
	return key
}

func extractServiceSeedFields(specSchema map[string]any) []models.FieldDefinition {
	properties, _ := specSchema["properties"].(map[string]any)
	if len(properties) == 0 {
		return nil
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	seeds := make([]models.FieldDefinition, 0, len(keys))
	for _, key := range keys {
		node, _ := properties[key].(map[string]any)
		if !isServiceLikeNode(key, node) {
			continue
		}

		seed, ok := buildServiceSeedField(key, node)
		if !ok {
			continue
		}
		seeds = append(seeds, seed)
	}

	return dedupeFields(seeds)
}

func isServiceLikeNode(name string, node map[string]any) bool {
	if node == nil {
		return false
	}
	nodeType := asString(node["type"])
	if nodeType != "" && nodeType != "object" {
		return false
	}
	properties, _ := node["properties"].(map[string]any)
	if len(properties) == 0 {
		return false
	}

	if _, ok := properties["configMaps"]; ok {
		return true
	}
	if _, ok := properties["envFromSecret"]; ok {
		return true
	}
	if _, ok := properties["spec"]; ok {
		// Typical component/service shape in operator CRDs.
		lower := strings.ToLower(name)
		if strings.Contains(lower, "madara") ||
			strings.Contains(lower, "bootstrapper") ||
			strings.Contains(lower, "orchestrator") ||
			strings.Contains(lower, "path") ||
			strings.Contains(lower, "dna") ||
			strings.Contains(lower, "faucet") {
			return true
		}
	}

	return false
}

func buildServiceSeedField(name string, node map[string]any) (models.FieldDefinition, bool) {
	properties, _ := node["properties"].(map[string]any)
	if len(properties) == 0 {
		return models.FieldDefinition{}, false
	}

	description := fmt.Sprintf("Service component '%s'. Expand as needed with optional fields.", name)
	prefix := "spec." + name

	if field, ok := seedFromConfigMaps(prefix, properties["configMaps"], description); ok {
		return field, true
	}
	if field, ok := seedFromEnvFromSecret(prefix, properties["envFromSecret"], description); ok {
		return field, true
	}
	if field, ok := seedFromSpecNode(prefix, properties["spec"], description); ok {
		return field, true
	}
	return models.FieldDefinition{}, false
}

func seedFromConfigMaps(prefix string, raw any, description string) (models.FieldDefinition, bool) {
	configMaps, _ := raw.(map[string]any)
	if configMaps == nil || asString(configMaps["type"]) != "array" {
		return models.FieldDefinition{}, false
	}

	items, _ := configMaps["items"].(map[string]any)
	itemProps, _ := items["properties"].(map[string]any)
	if len(itemProps) == 0 {
		return models.FieldDefinition{}, false
	}

	if _, ok := itemProps["name"]; ok {
		return models.FieldDefinition{
			Path:        prefix + ".configMaps[0].name",
			Description: description,
		}, true
	}
	if _, ok := itemProps["mountPath"]; ok {
		return models.FieldDefinition{
			Path:        prefix + ".configMaps[0].mountPath",
			Description: description,
		}, true
	}

	keys := make([]string, 0, len(itemProps))
	for key := range itemProps {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return models.FieldDefinition{
		Path:        prefix + ".configMaps[0]." + keys[0],
		Description: description,
	}, true
}

func seedFromEnvFromSecret(prefix string, raw any, description string) (models.FieldDefinition, bool) {
	envFromSecret, _ := raw.(map[string]any)
	if envFromSecret == nil || asString(envFromSecret["type"]) != "object" {
		return models.FieldDefinition{}, false
	}
	props, _ := envFromSecret["properties"].(map[string]any)
	if len(props) == 0 {
		return models.FieldDefinition{}, false
	}
	if _, ok := props["name"]; ok {
		return models.FieldDefinition{
			Path:        prefix + ".envFromSecret.name",
			Description: description,
		}, true
	}

	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return models.FieldDefinition{
		Path:        prefix + ".envFromSecret." + keys[0],
		Description: description,
	}, true
}

func seedFromSpecNode(prefix string, raw any, description string) (models.FieldDefinition, bool) {
	specNode, _ := raw.(map[string]any)
	if specNode == nil || asString(specNode["type"]) != "object" {
		return models.FieldDefinition{}, false
	}

	props, _ := specNode["properties"].(map[string]any)
	if len(props) == 0 {
		return models.FieldDefinition{}, false
	}

	preferred := []string{"roleArn", "serviceAccountName", "minUnavailable"}
	for _, key := range preferred {
		property, _ := props[key].(map[string]any)
		if property == nil {
			continue
		}
		defaultValue, _ := inferSchemaDefault(property)
		return models.FieldDefinition{
			Path:        prefix + ".spec." + key,
			Type:        normalizeFieldType(asString(property["type"])),
			Value:       defaultValue,
			Description: description,
		}, true
	}

	return firstScalarSeed(prefix+".spec", props, description, 0, 2)
}

func firstScalarSeed(
	prefix string,
	properties map[string]any,
	description string,
	depth int,
	maxDepth int,
) (models.FieldDefinition, bool) {
	if len(properties) == 0 || depth > maxDepth {
		return models.FieldDefinition{}, false
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		node, _ := properties[key].(map[string]any)
		if node == nil {
			continue
		}

		path := prefix + "." + key
		nodeType := asString(node["type"])
		if nodeType == "array" {
			items, _ := node["items"].(map[string]any)
			itemProps, _ := items["properties"].(map[string]any)
			if len(itemProps) > 0 {
				return firstScalarSeed(path+"[0]", itemProps, description, depth+1, maxDepth)
			}
			return models.FieldDefinition{
				Path:        path + "[0]",
				Description: description,
			}, true
		}

		nestedProps, _ := node["properties"].(map[string]any)
		if len(nestedProps) > 0 {
			if seed, ok := firstScalarSeed(path, nestedProps, description, depth+1, maxDepth); ok {
				return seed, true
			}
			continue
		}

		if nodeType == "object" || (nodeType == "" && node["additionalProperties"] != nil) {
			if seed, ok := mapSeedField(path, node, description); ok {
				return seed, true
			}
			continue
		}

		defaultValue, _ := inferSchemaDefault(node)
		return models.FieldDefinition{
			Path:        path,
			Type:        normalizeFieldType(nodeType),
			Value:       defaultValue,
			Description: description,
		}, true
	}

	return models.FieldDefinition{}, false
}

func parseWithRegexFallback(raw string) models.TemplateDefinition {
	kind := firstCapture(regexKind, raw)
	if kind == "" {
		kind = "CustomResource"
	}

	group := firstCapture(regexGroup, raw)
	version := firstCapture(regexVersionRef, raw)
	if version == "" {
		version = firstCapture(regexVersion, raw)
	}

	apiVersion := "example.io/v1"
	if group != "" && version != "" {
		apiVersion = fmt.Sprintf("%s/%s", group, version)
	}

	fields := make([]models.FieldDefinition, 0, 8)
	for _, match := range regexField.FindAllStringSubmatch(raw, -1) {
		name := match[1]
		if strings.EqualFold(name, "type") ||
			strings.EqualFold(name, "properties") ||
			strings.EqualFold(name, "items") ||
			strings.EqualFold(name, "description") ||
			strings.EqualFold(name, "required") ||
			strings.EqualFold(name, "metadata") ||
			strings.EqualFold(name, "spec") ||
			strings.EqualFold(name, "status") {
			continue
		}
		path := "spec." + name
		if fieldPathExists(fields, path) {
			continue
		}
		fields = append(fields, models.FieldDefinition{
			Path:        path,
			Description: fmt.Sprintf("Inferred from parsed field '%s'.", name),
		})
		if len(fields) >= 8 {
			break
		}
	}

	if len(fields) == 0 {
		fields = append(fields, models.FieldDefinition{
			Path:        "spec.example",
			Description: "No schema fields inferred from input. Replace this with real fields.",
		})
	}

	return models.TemplateDefinition{
		ID:         normalizeID("parsed-" + kind),
		Title:      kind + " (Parsed)",
		APIVersion: apiVersion,
		Kind:       kind,
		Note:       "Generated with fallback parser. Prefer full CRD YAML for richer field inference.",
		DefaultFields: append([]models.FieldDefinition{
			{Path: "metadata.name", Value: strings.ToLower(kind) + "-sample", Description: "Name for this resource."},
			{Path: "metadata.namespace", Value: "default", Description: "Namespace for this resource."},
		}, fields...),
		OptionalFields: []models.FieldDefinition{
			{Path: "metadata.labels.app", Description: "Optional labels."},
			{Path: "metadata.annotations.owner", Description: "Optional annotations."},
		},
	}
}

func decodeYAMLDocuments(raw string) ([]map[string]any, error) {
	decoder := yaml.NewDecoder(strings.NewReader(raw))
	docs := make([]map[string]any, 0, 4)

	for {
		var decoded any
		err := decoder.Decode(&decoded)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		docMap, _ := decoded.(map[string]any)
		if len(docMap) == 0 {
			continue
		}
		docs = append(docs, docMap)
	}

	if len(docs) == 0 {
		return nil, errors.New("no YAML documents found")
	}

	return docs, nil
}

func selectPrimaryResourceDoc(docs []map[string]any) (map[string]any, bool) {
	for _, doc := range docs {
		kind := asString(doc["kind"])
		if strings.EqualFold(kind, "CustomResourceDefinition") {
			return doc, true
		}
	}

	for _, doc := range docs {
		kind := asString(doc["kind"])
		if !strings.EqualFold(kind, "List") {
			continue
		}
		items, _ := doc["items"].([]any)
		for _, item := range items {
			resourceMap, _ := item.(map[string]any)
			if resourceMap == nil {
				continue
			}
			if strings.EqualFold(asString(resourceMap["kind"]), "CustomResourceDefinition") {
				return resourceMap, true
			}
		}
	}

	for _, doc := range docs {
		if asString(doc["kind"]) != "" {
			return doc, true
		}
	}

	if len(docs) > 0 {
		return docs[0], true
	}
	return nil, false
}

func nested(root map[string]any, keys ...string) any {
	var current any = root
	for _, key := range keys {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		next, ok := obj[key]
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

func nestedWithArrayPath(root map[string]any, keys ...string) any {
	var current any = root
	for _, key := range keys {
		if strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]") {
			arr, ok := current.([]any)
			if !ok {
				return nil
			}
			indexText := strings.TrimSuffix(strings.TrimPrefix(key, "["), "]")
			index := 0
			_, _ = fmt.Sscanf(indexText, "%d", &index)
			if index < 0 || index >= len(arr) {
				return nil
			}
			current = arr[index]
			continue
		}
		obj, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		next, ok := obj[key]
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

func firstVersion(candidate any) string {
	versions, ok := candidate.([]any)
	if !ok {
		return ""
	}
	for _, version := range versions {
		verMap, ok := version.(map[string]any)
		if !ok {
			continue
		}
		name := asString(verMap["name"])
		if name != "" {
			return name
		}
	}
	return ""
}

func hasCRDSchema(root map[string]any) bool {
	legacy := nested(root, "spec", "validation", "openAPIV3Schema")
	if legacy != nil {
		return true
	}

	versions, _ := nested(root, "spec", "versions").([]any)
	for _, item := range versions {
		versionMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		schema, _ := versionMap["schema"].(map[string]any)
		if schema == nil {
			continue
		}
		if schema["openAPIV3Schema"] != nil {
			return true
		}
	}
	return false
}

func asString(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func asBool(value any) bool {
	typed, ok := value.(bool)
	if !ok {
		return false
	}
	return typed
}

func normalizeFieldType(value string) string {
	switch value {
	case "integer", "number":
		return "number"
	case "boolean":
		return "boolean"
	default:
		return ""
	}
}

func firstCapture(pattern *regexp.Regexp, input string) string {
	match := pattern.FindStringSubmatch(input)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func fieldPathExists(fields []models.FieldDefinition, path string) bool {
	for _, field := range fields {
		if field.Path == path {
			return true
		}
	}
	return false
}

func normalizeID(input string) string {
	lower := strings.ToLower(input)
	lower = strings.ReplaceAll(lower, "_", "-")
	lower = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(lower, "")
	lower = strings.Trim(lower, "-")
	if lower == "" {
		return "parsed-custom-resource"
	}
	return lower
}
