package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptCacheDiagnosticsMigrationStoresHashesOnly(t *testing.T) {
	content, err := FS.ReadFile("194_usage_log_prompt_cache_diagnostics.sql")
	require.NoError(t, err)

	sql := strings.ToLower(strings.Join(strings.Fields(string(content)), " "))
	for _, column := range []string{
		"prompt_cache_key_hash",
		"prompt_cache_key_source",
		"prompt_cache_prefix_hash",
		"prompt_cache_tools_hash",
		"prompt_cache_system_hash",
	} {
		require.Contains(t, sql, "add column if not exists "+column)
	}
	require.NotContains(t, sql, "prompt_cache_key varchar")
	require.NotContains(t, sql, "prompt_body")
}
