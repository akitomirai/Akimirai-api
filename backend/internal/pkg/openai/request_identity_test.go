package openai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPairCodexClientIdentity(t *testing.T) {
	tests := []struct {
		name           string
		ua             string
		wantOriginator string
		wantUA         string
		wantOK         bool
	}{
		{name: "cli identity", ua: "codex_cli_rs/0.144.1 (Ubuntu 22.4.0; x86_64)", wantOriginator: "codex_cli_rs", wantUA: "codex_cli_rs/0.144.1 (Ubuntu 22.4.0; x86_64)", wantOK: true},
		{name: "tui identity", ua: "codex-tui/0.140.2 (Mac OS X 14.0; arm64) iTerm (codex-tui; 0.140.2)", wantOriginator: "codex-tui", wantUA: "codex-tui/0.140.2 (Mac OS X 14.0; arm64) iTerm (codex-tui; 0.140.2)", wantOK: true},
		{name: "desktop family preserves case", ua: "Codex Desktop/1.2.3", wantOriginator: "Codex Desktop", wantUA: "Codex Desktop/1.2.3", wantOK: true},
		{name: "override trailer restores identity", ua: "cccc/0.142.0 (Ubuntu 22.4.0; x86_64) screen (codex-tui; 0.142.0)", wantOriginator: "codex-tui", wantUA: "codex-tui/0.142.0 (Ubuntu 22.4.0; x86_64) screen (codex-tui; 0.142.0)", wantOK: true},
		{name: "canonicalizes exact identity", ua: "CODEX_CLI_RS/1.0.0", wantOriginator: "codex_cli_rs", wantUA: "codex_cli_rs/1.0.0", wantOK: true},
		{name: "rejects slash in trailer", ua: "foo/1.0 (Codex Desktop/2; 1.0)"},
		{name: "rejects non ascii", ua: "Codex \xc3\xa9vil/1.0.0"},
		{name: "rejects overlong identity", ua: "Codex " + strings.Repeat("a", 80) + "/1.0.0"},
		{name: "rejects third party", ua: "luna/1.0.0"},
		{name: "rejects missing slash", ua: "curl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originator, pairedUA, ok := PairCodexClientIdentity(tt.ua)
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantOriginator, originator)
			require.Equal(t, tt.wantUA, pairedUA)
		})
	}
}
