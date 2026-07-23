package service

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestInspectOpenAIPromptCacheDiagnosticsStableAcrossAppendOnlyTurns(t *testing.T) {
	base := []byte(`{"model":"gpt-5.6-sol","instructions":"be concise","tools":[{"type":"function","name":"lookup"}],"input":[{"role":"user","content":"Question A"}]}`)
	extended := []byte(`{"model":"gpt-5.6-sol","instructions":"be concise","tools":[{"type":"function","name":"lookup"}],"input":[{"role":"user","content":"Question A"},{"role":"assistant","content":"Answer"},{"role":"user","content":"Question B"}]}`)

	first := InspectOpenAIPromptCacheDiagnostics(base, "session-1", PromptCacheKeySourceClientBody)
	second := InspectOpenAIPromptCacheDiagnostics(extended, "session-1", PromptCacheKeySourceClientBody)

	require.Equal(t, first, second)
	require.Len(t, first.KeyHash, 64)
	require.Len(t, first.PrefixHash, 64)
	require.Len(t, first.ToolsHash, 64)
	require.Len(t, first.SystemHash, 64)
	require.NotContains(t, first.KeyHash, "session-1")
}

func TestInspectOpenAIPromptCacheDiagnosticsChangesOwningComponent(t *testing.T) {
	base := []byte(`{"model":"gpt-5.6-sol","instructions":"be concise","tools":[{"type":"function","name":"lookup"}],"input":[{"role":"user","content":"Question A"}]}`)
	differentTools := []byte(`{"model":"gpt-5.6-sol","instructions":"be concise","tools":[{"type":"function","name":"save"}],"input":[{"role":"user","content":"Question A"}]}`)
	differentInstructions := []byte(`{"model":"gpt-5.6-sol","instructions":"be detailed","tools":[{"type":"function","name":"lookup"}],"input":[{"role":"user","content":"Question A"}]}`)

	one := InspectOpenAIPromptCacheDiagnostics(base, "session-1", PromptCacheKeySourceClientBody)
	tools := InspectOpenAIPromptCacheDiagnostics(differentTools, "session-1", PromptCacheKeySourceClientBody)
	instructions := InspectOpenAIPromptCacheDiagnostics(differentInstructions, "session-1", PromptCacheKeySourceClientBody)

	require.NotEqual(t, one.ToolsHash, tools.ToolsHash)
	require.Equal(t, one.SystemHash, tools.SystemHash)
	require.NotEqual(t, one.SystemHash, instructions.SystemHash)
	require.NotEqual(t, one.PrefixHash, tools.PrefixHash)
}

func TestInspectOpenAIPromptCacheDiagnosticsInstructionsDoNotHideInputDeveloperMessages(t *testing.T) {
	first := InspectOpenAIPromptCacheDiagnostics(
		[]byte(`{"model":"gpt-5.6-sol","instructions":"base","input":[{"role":"developer","content":"policy-a"},{"role":"user","content":"hi"}]}`),
		"session-1",
		PromptCacheKeySourceClientBody,
	)
	second := InspectOpenAIPromptCacheDiagnostics(
		[]byte(`{"model":"gpt-5.6-sol","instructions":"base","input":[{"role":"developer","content":"policy-b"},{"role":"user","content":"hi"}]}`),
		"session-1",
		PromptCacheKeySourceClientBody,
	)

	require.NotEqual(t, first.SystemHash, second.SystemHash)
	require.NotEqual(t, first.PrefixHash, second.PrefixHash)
}

func TestPromptCacheDiagnosticsForRequestClassifiesIdentitySource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	c.Request.Header.Set("session_id", "header-session")

	header := PromptCacheDiagnosticsForRequest(c, []byte(`{"model":"gpt-5.6-sol","input":"hi"}`), "header-session", false)
	body := PromptCacheDiagnosticsForRequest(nil, []byte(`{"model":"gpt-5.6-sol","prompt_cache_key":"body-session","input":"hi"}`), "body-session", false)
	compat := PromptCacheDiagnosticsForRequest(nil, []byte(`{"model":"gpt-5.6-sol","input":"hi"}`), "derived-session", true)

	require.Equal(t, PromptCacheKeySourceClientHeader, header.Source)
	require.Equal(t, PromptCacheKeySourceClientBody, body.Source)
	require.Equal(t, PromptCacheKeySourceCompatDerived, compat.Source)
}

func TestPromptCacheDiagnosticsForRequestReportsBodyKeyWhenHeaderAlsoExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	c.Request.Header.Set("session_id", "header-session")

	got := PromptCacheDiagnosticsForRequest(
		c,
		[]byte(`{"model":"gpt-5.6-sol","prompt_cache_key":"body-session","input":"hi"}`),
		"header-session",
		false,
	)

	require.Equal(t, PromptCacheKeySourceClientBody, got.Source)
	require.Equal(t, hashPromptCacheDiagnosticValue("body-session"), got.KeyHash)
	require.NotEqual(t, hashPromptCacheDiagnosticValue("header-session"), got.KeyHash)
}

func TestInspectOpenAIPromptCacheDiagnosticsDoesNotCreateIdentityFromModelOnly(t *testing.T) {
	got := InspectOpenAIPromptCacheDiagnostics([]byte(`{"model":"gpt-5.6-sol"}`), "", PromptCacheKeySourceNone)
	require.Equal(t, PromptCacheKeySourceNone, got.Source)
	require.Empty(t, got.KeyHash)
	require.Empty(t, got.PrefixHash)
	require.Empty(t, got.ToolsHash)
	require.Empty(t, got.SystemHash)
}

func TestEnsureOpenAIPromptCacheKeyPromotesExplicitHeaderIdentity(t *testing.T) {
	body := []byte(`{"model":"gpt-5.6-sol","input":"hi"}`)
	got, changed, err := EnsureOpenAIPromptCacheKey(body, " session-1 ")
	require.NoError(t, err)
	require.True(t, changed)
	require.JSONEq(t, `{"model":"gpt-5.6-sol","input":"hi","prompt_cache_key":"session-1"}`, string(got))

	existing, changed, err := EnsureOpenAIPromptCacheKey([]byte(`{"prompt_cache_key":"client-key"}`), "header-key")
	require.NoError(t, err)
	require.False(t, changed)
	require.JSONEq(t, `{"prompt_cache_key":"client-key"}`, string(existing))
}
