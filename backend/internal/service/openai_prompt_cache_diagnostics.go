package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// PromptCacheKeySource is intentionally bounded because it is persisted in
// usage_logs and rendered in the admin diagnostics UI.
type PromptCacheKeySource string

const (
	PromptCacheKeySourceNone          PromptCacheKeySource = "none"
	PromptCacheKeySourceClientHeader  PromptCacheKeySource = "client_header"
	PromptCacheKeySourceClientBody    PromptCacheKeySource = "client_body"
	PromptCacheKeySourceCompatDerived PromptCacheKeySource = "compat_derived"
)

// PromptCacheDiagnostics contains hashes only. Raw prompts and raw cache keys
// must never cross the request diagnostics boundary.
type PromptCacheDiagnostics struct {
	KeyHash    string
	Source     PromptCacheKeySource
	PrefixHash string
	ToolsHash  string
	SystemHash string
}

// EnsureOpenAIPromptCacheKey promotes an explicit conversation identity into
// the Responses body only when the client did not already provide a cache key.
// Callers must scope this to an upstream known to support prompt_cache_key.
func EnsureOpenAIPromptCacheKey(body []byte, key string) ([]byte, bool, error) {
	key = strings.TrimSpace(key)
	if key == "" || strings.TrimSpace(gjson.GetBytes(body, "prompt_cache_key").String()) != "" {
		return body, false, nil
	}
	out, err := sjson.SetBytes(body, "prompt_cache_key", key)
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// InspectOpenAIPromptCacheDiagnostics hashes the effective cache identity and
// the reusable prompt-prefix components. The later conversation turns are
// deliberately excluded from these component hashes.
func InspectOpenAIPromptCacheDiagnostics(body []byte, effectiveKey string, source PromptCacheKeySource) PromptCacheDiagnostics {
	result := PromptCacheDiagnostics{Source: normalizePromptCacheKeySource(source)}
	if key := strings.TrimSpace(effectiveKey); key != "" {
		result.KeyHash = hashPromptCacheDiagnosticValue(key)
	}
	if seed := deriveOpenAIStablePrefixSessionSeed(body); seed != "" {
		result.PrefixHash = hashPromptCacheDiagnosticValue(seed)
	}
	result.ToolsHash = hashPromptCacheTools(body)
	result.SystemHash = hashPromptCacheSystem(body)
	return result
}

// PromptCacheDiagnosticsForRequest resolves the client-visible identity
// source using the same priority as OpenAI session routing. compatDerived is
// set by the Chat Completions compatibility bridge after it creates a key.
func PromptCacheDiagnosticsForRequest(c *gin.Context, body []byte, effectiveKey string, compatDerived bool) PromptCacheDiagnostics {
	if compatDerived {
		return InspectOpenAIPromptCacheDiagnostics(body, effectiveKey, PromptCacheKeySourceCompatDerived)
	}

	// An existing body key wins because EnsureOpenAIPromptCacheKey deliberately
	// preserves it. Header identity is promoted only when the body has no key.
	if bodyKey := strings.TrimSpace(gjson.GetBytes(body, "prompt_cache_key").String()); bodyKey != "" {
		return InspectOpenAIPromptCacheDiagnostics(body, bodyKey, PromptCacheKeySourceClientBody)
	}
	if c != nil && (strings.TrimSpace(c.GetHeader("session_id")) != "" || strings.TrimSpace(c.GetHeader("conversation_id")) != "") {
		return InspectOpenAIPromptCacheDiagnostics(body, effectiveKey, PromptCacheKeySourceClientHeader)
	}
	if strings.TrimSpace(effectiveKey) != "" {
		// A caller may provide a protocol-specific explicit key without exposing
		// it as a generic header/body field. Keep it classified as client input.
		return InspectOpenAIPromptCacheDiagnostics(body, effectiveKey, PromptCacheKeySourceClientBody)
	}
	return InspectOpenAIPromptCacheDiagnostics(body, "", PromptCacheKeySourceNone)
}

func normalizePromptCacheKeySource(source PromptCacheKeySource) PromptCacheKeySource {
	switch source {
	case PromptCacheKeySourceClientHeader, PromptCacheKeySourceClientBody, PromptCacheKeySourceCompatDerived:
		return source
	default:
		return PromptCacheKeySourceNone
	}
}

func hashPromptCacheTools(body []byte) string {
	parts := make([]string, 0, 2)
	for _, field := range []string{"tools", "functions"} {
		value := gjson.GetBytes(body, field)
		if normalized, ok := normalizeNonEmptyCompatSeedJSON(value); ok {
			parts = append(parts, field+"="+normalized)
		}
	}
	return hashPromptCacheParts(parts)
}

func hashPromptCacheSystem(body []byte) string {
	parts := make([]string, 0, 4)
	if instructions := gjson.GetBytes(body, "instructions"); strings.TrimSpace(instructions.String()) != "" {
		parts = append(parts, "instructions="+instructions.String())
	}
	appendSystem := func(items gjson.Result) {
		if !items.Exists() || !items.IsArray() {
			return
		}
		items.ForEach(func(_, item gjson.Result) bool {
			role := strings.TrimSpace(item.Get("role").String())
			if role != "system" && role != "developer" {
				return true
			}
			if normalized, ok := normalizeNonEmptyCompatSeedJSON(item.Get("content")); ok {
				parts = append(parts, role+"="+normalized)
			}
			return true
		})
	}
	if messages := gjson.GetBytes(body, "messages"); messages.Exists() && messages.IsArray() {
		appendSystem(messages)
	} else {
		appendSystem(gjson.GetBytes(body, "input"))
	}
	return hashPromptCacheParts(parts)
}

func hashPromptCacheParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return hashPromptCacheDiagnosticValue(strings.Join(parts, "\x00"))
}

func hashPromptCacheDiagnosticValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
