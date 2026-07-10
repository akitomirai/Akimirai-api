package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAPIKeyAllowedModelsTrimsDeduplicatesAndDropsEmptyValues(t *testing.T) {
	models, err := NormalizeAPIKeyAllowedModels([]string{" gpt-5.6-sol ", "", "gpt-5.6-sol", "gemini-2.5-pro"})
	require.NoError(t, err)
	require.Equal(t, []string{"gpt-5.6-sol", "gemini-2.5-pro"}, models)
}

func TestNormalizeAPIKeyAllowedModelsAppliesLimitsAfterNormalization(t *testing.T) {
	duplicates := make([]string, apiKeyMaxAllowedModels+1)
	for i := range duplicates {
		duplicates[i] = "gpt-5.6-sol"
	}
	models, err := NormalizeAPIKeyAllowedModels(duplicates)
	require.NoError(t, err)
	require.Equal(t, []string{"gpt-5.6-sol"}, models)

	unique := make([]string, apiKeyMaxAllowedModels+1)
	for i := range unique {
		unique[i] = fmt.Sprintf("model-%03d", i)
	}
	_, err = NormalizeAPIKeyAllowedModels(unique)
	require.ErrorIs(t, err, ErrAPIKeyAllowedModelsTooMany)

	_, err = NormalizeAPIKeyAllowedModels([]string{strings.Repeat("m", apiKeyMaxModelNameLen+1)})
	require.ErrorIs(t, err, ErrAPIKeyAllowedModelTooLong)
}

func TestAPIKeyAllowsModelUsesExactMatchAndEmptyListIsUnrestricted(t *testing.T) {
	require.True(t, (&APIKey{}).AllowsModel("gpt-5.6-sol"))

	key := &APIKey{AllowedModels: []string{"gpt-5.6-sol", "gemini-2.5-pro"}}
	require.True(t, key.AllowsModel(" gpt-5.6-sol "))
	require.False(t, key.AllowsModel("gpt-5.6"))
	require.False(t, key.AllowsModel(""))
	require.Equal(t, []string{"gemini-2.5-pro"}, key.FilterAllowedModels([]string{"gpt-5.5", "gemini-2.5-pro"}))
}
