package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func filterCodexModelsManifest(body []byte, apiKey *service.APIKey) ([]byte, error) {
	return filterJSONModelArray(body, "models", "slug", func(model string) string {
		return strings.TrimSpace(model)
	}, apiKey)
}

func filterGeminiModelsValue(value any, apiKey *service.APIKey) ([]byte, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal Gemini models response: %w", err)
	}
	return filterGeminiModelsBody(body, apiKey)
}

func filterGeminiModelsBody(body []byte, apiKey *service.APIKey) ([]byte, error) {
	return filterJSONModelArray(body, "models", "name", func(model string) string {
		return strings.TrimPrefix(strings.TrimSpace(model), "models/")
	}, apiKey)
}

func filterJSONModelArray(body []byte, arrayField, idField string, normalize func(string) string, apiKey *service.APIKey) ([]byte, error) {
	if apiKey == nil || len(apiKey.AllowedModels) == 0 {
		return body, nil
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("decode model list: %w", err)
	}
	rawModels, ok := root[arrayField]
	if !ok {
		return nil, fmt.Errorf("model list is missing %q", arrayField)
	}

	var models []json.RawMessage
	if err := json.Unmarshal(rawModels, &models); err != nil {
		return nil, fmt.Errorf("decode model entries: %w", err)
	}
	filtered := make([]json.RawMessage, 0, len(models))
	for _, rawModel := range models {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(rawModel, &fields); err != nil {
			continue
		}
		var modelID string
		if err := json.Unmarshal(fields[idField], &modelID); err != nil {
			continue
		}
		if apiKey.AllowsModel(normalize(modelID)) {
			filtered = append(filtered, rawModel)
		}
	}

	filteredJSON, err := json.Marshal(filtered)
	if err != nil {
		return nil, fmt.Errorf("encode filtered model entries: %w", err)
	}
	root[arrayField] = filteredJSON
	result, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("encode filtered model list: %w", err)
	}
	return result, nil
}
