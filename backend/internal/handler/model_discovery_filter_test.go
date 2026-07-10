package handler

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestFilterCodexModelsManifestPreservesAllowedEntriesAndMetadata(t *testing.T) {
	body := []byte(`{"models":[{"slug":"gpt-5.6-sol","display_name":"SOL"},{"slug":"gpt-5.5"}],"client_version":"0.137.0"}`)

	filtered, err := filterCodexModelsManifest(body, &service.APIKey{AllowedModels: []string{"gpt-5.6-sol"}})
	require.NoError(t, err)

	var got struct {
		Models []struct {
			Slug        string `json:"slug"`
			DisplayName string `json:"display_name"`
		} `json:"models"`
		ClientVersion string `json:"client_version"`
	}
	require.NoError(t, json.Unmarshal(filtered, &got))
	require.Equal(t, "0.137.0", got.ClientVersion)
	require.Len(t, got.Models, 1)
	require.Equal(t, []string{"gpt-5.6-sol"}, []string{got.Models[0].Slug})
	require.Equal(t, "SOL", got.Models[0].DisplayName)
}

func TestFilterGeminiModelsBodyNormalizesModelsPrefix(t *testing.T) {
	body := []byte(`{"models":[{"name":"models/gemini-2.5-pro"},{"name":"models/gemini-2.0-flash"}],"nextPageToken":"next"}`)

	filtered, err := filterGeminiModelsBody(body, &service.APIKey{AllowedModels: []string{"gemini-2.5-pro"}})
	require.NoError(t, err)
	require.JSONEq(t, `{"models":[{"name":"models/gemini-2.5-pro"}],"nextPageToken":"next"}`, string(filtered))
}

func TestFilterJSONModelArrayLeavesUnrestrictedBodyUntouched(t *testing.T) {
	body := []byte("{ \"models\": [] }")

	filtered, err := filterCodexModelsManifest(body, &service.APIKey{})
	require.NoError(t, err)
	require.Equal(t, body, filtered)
}
