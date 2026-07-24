package service

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/google/uuid"
)

const codexUpstreamMinVersion = "0.144.0"

// ensureCodexIdentityHeaders fills the complete OAuth ChatGPT Codex identity.
// Existing official User-Agent and version values are retained; enforce then
// pairs the identity and raises an explicitly stale version when necessary.
func ensureCodexIdentityHeaders(headers http.Header) {
	if headers == nil {
		return
	}
	if strings.TrimSpace(headers.Get("user-agent")) == "" {
		headers.Set("user-agent", codexCLIUserAgent)
	}
	if strings.TrimSpace(headers.Get("originator")) == "" {
		headers.Set("originator", "codex_cli_rs")
	}
	if strings.TrimSpace(headers.Get("version")) == "" {
		headers.Set("version", codexCLIVersion)
	}
	headers.Set("OpenAI-Beta", "responses=experimental")
}

// applyOpenAICodexProbeHeaders fills synthetic probe identity and gives every
// probe a distinct engine fingerprint.
func applyOpenAICodexProbeHeaders(headers http.Header) {
	if headers == nil {
		return
	}
	ensureCodexIdentityHeaders(headers)
	headers.Set("X-Codex-Window-ID", uuid.NewString())
}

// enforceCodexIdentityHeaders pairs originator with the final outbound
// User-Agent and raises an explicitly supplied stale version header.
// Requests without originator remain untouched unless their caller explicitly
// restores the complete identity with ensureCodexIdentityHeaders first.
func enforceCodexIdentityHeaders(headers http.Header) {
	if headers == nil || headers.Get("originator") == "" {
		return
	}
	originator, pairedUA, ok := openai.PairCodexClientIdentity(headers.Get("user-agent"))
	if !ok {
		originator, pairedUA = "codex_cli_rs", codexCLIUserAgent
	}
	headers.Set("user-agent", pairedUA)
	headers.Set("originator", originator)
	if version := strings.TrimSpace(headers.Get("version")); version != "" && CompareVersions(version, codexUpstreamMinVersion) < 0 {
		headers.Set("version", codexCLIVersion)
	}
}
