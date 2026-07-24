package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func requireOpenAICodexProbeHeaders(t *testing.T, h http.Header) {
	t.Helper()
	require.Equal(t, codexCLIUserAgent, h.Get("User-Agent"))
	require.Equal(t, "codex_cli_rs", h.Get("Originator"))
	require.Equal(t, codexCLIVersion, h.Get("Version"))
	require.Equal(t, "responses=experimental", h.Get("OpenAI-Beta"))
	require.NotEmpty(t, h.Get("X-Codex-Window-ID"))
}

func TestEnsureCodexIdentityHeaders(t *testing.T) {
	t.Run("fills missing identity", func(t *testing.T) {
		headers := make(http.Header)

		ensureCodexIdentityHeaders(headers)
		enforceCodexIdentityHeaders(headers)

		require.Equal(t, "codex_cli_rs", headers.Get("originator"))
		require.Equal(t, codexCLIUserAgent, headers.Get("user-agent"))
		require.Equal(t, codexCLIVersion, headers.Get("version"))
		require.Equal(t, "responses=experimental", headers.Get("OpenAI-Beta"))
	})

	t.Run("keeps official user agent and accepted version", func(t *testing.T) {
		const tuiUA = "codex-tui/9.9.9 (Mac OS X 14.0; arm64) iTerm (codex-tui; 9.9.9)"
		headers := make(http.Header)
		headers.Set("user-agent", tuiUA)
		headers.Set("version", "9.9.9")
		headers.Set("OpenAI-Beta", "assistants=v2")

		ensureCodexIdentityHeaders(headers)
		enforceCodexIdentityHeaders(headers)

		require.Equal(t, "codex-tui", headers.Get("originator"))
		require.Equal(t, tuiUA, headers.Get("user-agent"))
		require.Equal(t, "9.9.9", headers.Get("version"))
		require.Equal(t, "responses=experimental", headers.Get("OpenAI-Beta"))
	})
}

func TestEnforceCodexIdentityHeaders(t *testing.T) {
	const tuiUA = "codex-tui/0.140.2 (Mac OS X 14.0; arm64) iTerm (codex-tui; 0.140.2)"
	tests := []struct {
		name, originator, userAgent, version string
		wantOriginator, wantUA, wantVersion  string
	}{
		{name: "repairs mismatched originator", originator: "codex_cli_rs", userAgent: tuiUA, wantOriginator: "codex-tui", wantUA: tuiUA},
		{name: "keeps paired official identity", originator: "codex-tui", userAgent: tuiUA, wantOriginator: "codex-tui", wantUA: tuiUA},
		{name: "falls back from third party ua", originator: "opencode", userAgent: "luna/1.0.0", wantOriginator: "codex_cli_rs", wantUA: codexCLIUserAgent},
		{name: "raises stale version", originator: "codex_cli_rs", userAgent: "codex_cli_rs/0.125.0", version: "0.125.0", wantOriginator: "codex_cli_rs", wantUA: "codex_cli_rs/0.125.0", wantVersion: codexCLIVersion},
		{name: "keeps accepted version", originator: "codex_cli_rs", userAgent: "codex_cli_rs/0.145.0", version: "0.145.0", wantOriginator: "codex_cli_rs", wantUA: "codex_cli_rs/0.145.0", wantVersion: "0.145.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := make(http.Header)
			headers.Set("originator", tt.originator)
			headers.Set("user-agent", tt.userAgent)
			if tt.version != "" {
				headers.Set("version", tt.version)
			}
			enforceCodexIdentityHeaders(headers)
			require.Equal(t, tt.wantOriginator, headers.Get("originator"))
			require.Equal(t, tt.wantUA, headers.Get("user-agent"))
			require.Equal(t, tt.wantVersion, headers.Get("version"))
		})
	}
}

func TestEnforceCodexIdentityHeaders_NoOriginatorIsNoop(t *testing.T) {
	headers := make(http.Header)
	headers.Set("user-agent", "third-party-client/1.0.0")
	enforceCodexIdentityHeaders(headers)
	require.Empty(t, headers.Get("originator"))
	require.Equal(t, "third-party-client/1.0.0", headers.Get("user-agent"))
}
