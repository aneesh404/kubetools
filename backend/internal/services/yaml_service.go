package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aneeshchawla/kubetools/backend/internal/models"
	"gopkg.in/yaml.v3"
)

var pathRegex = regexp.MustCompile(`([^\[]+)|\[(\d+)\]`)
var numberRegex = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

type YAMLService struct{}

func NewYAMLService() *YAMLService {
	return &YAMLService{}
}

func (s *YAMLService) GenerateYAML(apiVersion, kind string, fields []models.FieldDefinition) (string, error) {
	if strings.TrimSpace(apiVersion) == "" {
		return "", fmt.Errorf("apiVersion is required")
	}
	if strings.TrimSpace(kind) == "" {
		return "", fmt.Errorf("kind is required")
	}

	resource := map[string]any{
		"apiVersion": apiVersion,
		"kind":       kind,
	}

	for _, field := range fields {
		path := strings.TrimSpace(field.Path)
		if path == "" {
			continue
		}
		setValue(resource, parsePath(path), parseValue(field.Value, field.Type))
	}

	output, err := yaml.Marshal(resource)
	if err != nil {
		return "", fmt.Errorf("marshal YAML: %w", err)
	}
	return string(output), nil
}

func parsePath(path string) []any {
	segments := make([]any, 0)
	parts := strings.Split(path, ".")
	for _, part := range parts {
		matches := pathRegex.FindAllStringSubmatch(part, -1)
		for _, match := range matches {
			if match[1] != "" {
				segments = append(segments, match[1])
				continue
			}
			idx, err := strconv.Atoi(match[2])
			if err == nil {
				segments = append(segments, idx)
			}
		}
	}
	return segments
}

func parseValue(value string, valueType string) any {
	trimmed := strings.TrimSpace(value)
	if valueType == "number" || numberRegex.MatchString(trimmed) {
		floatValue, err := strconv.ParseFloat(trimmed, 64)
		if err == nil {
			if floatValue == float64(int64(floatValue)) {
				return int64(floatValue)
			}
			return floatValue
		}
	}
	if valueType == "boolean" || trimmed == "true" || trimmed == "false" {
		return trimmed == "true"
	}
	return value
}

func setValue(root map[string]any, segments []any, value any) {
	var node any = root
	setNode(&node, segments, value)
}

func setNode(node *any, segments []any, value any) {
	if len(segments) == 0 {
		return
	}

	segment := segments[0]
	last := len(segments) == 1

	switch key := segment.(type) {
	case string:
		var obj map[string]any
		if *node == nil {
			obj = map[string]any{}
			*node = obj
		} else {
			cast, ok := (*node).(map[string]any)
			if !ok {
				return
			}
			obj = cast
		}

		if last {
			obj[key] = value
			return
		}

		child := obj[key]
		setNode(&child, segments[1:], value)
		obj[key] = child
	case int:
		var arr []any
		if *node == nil {
			arr = make([]any, 0)
		} else {
			cast, ok := (*node).([]any)
			if !ok {
				return
			}
			arr = cast
		}

		for len(arr) <= key {
			arr = append(arr, nil)
		}

		if last {
			arr[key] = value
			*node = arr
			return
		}

		child := arr[key]
		setNode(&child, segments[1:], value)
		arr[key] = child
		*node = arr
	}
}
