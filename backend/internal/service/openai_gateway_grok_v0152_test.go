package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func newGrokV0152TestContext(apiKeyID int64, platform string) *gin.Context {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Set("api_key", &APIKey{ID: apiKeyID, Group: &Group{Platform: platform}})
	return c
}

func TestV0152GrokCacheIdentityStableAndTenantScoped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	round1 := []byte(`{"model":"grok","instructions":"be concise","input":[{"role":"user","content":"first"}]}`)
	round2 := []byte(`{"model":"grok","instructions":"be concise","input":[{"role":"user","content":"first"},{"role":"assistant","content":"answer"},{"role":"user","content":"second"}]}`)

	first := resolveGrokCacheIdentity(newGrokV0152TestContext(101, PlatformGrok), round1, "", "grok-4.5")
	second := resolveGrokCacheIdentity(newGrokV0152TestContext(101, PlatformGrok), round2, "", "grok-4.5")
	otherTenant := resolveGrokCacheIdentity(newGrokV0152TestContext(102, PlatformGrok), round1, "", "grok-4.5")
	otherModel := resolveGrokCacheIdentity(newGrokV0152TestContext(101, PlatformGrok), round1, "", "grok-4.3")

	require.NotEmpty(t, first)
	require.Len(t, first, 36)
	require.Equal(t, first, second)
	require.NotEqual(t, first, otherTenant)
	require.NotEqual(t, first, otherModel)
}

func TestV0152GrokConversationHeaderOnlyAffectsGrokScheduling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"grok","prompt_cache_key":"body-session","input":"hi"}`)

	grokContext := newGrokV0152TestContext(201, PlatformGrok)
	grokContext.Request.Header.Set(grokConversationIDHeader, "native-grok-session")
	require.Equal(t, "native-grok-session", (&OpenAIGatewayService{}).ExtractSessionID(grokContext, body))

	openAIContext := newGrokV0152TestContext(201, PlatformOpenAI)
	openAIContext.Request.Header.Set(grokConversationIDHeader, "must-be-ignored")
	require.Equal(t, "body-session", (&OpenAIGatewayService{}).ExtractSessionID(openAIContext, body))
}

func TestV0152ApplyGrokCacheIdentityPreservesExplicitToolIntent(t *testing.T) {
	intentBody := []byte(`{"model":"grok","tools":[{"type":"namespace","name":"client_tools"}],"tool_choice":{"type":"namespace","name":"client_tools"}}`)
	patchedBody := []byte(`{"model":"grok-4.5","input":"hello"}`)

	body, err := applyGrokResponsesCacheIdentity(patchedBody, intentBody, "isolated-id", true)

	require.NoError(t, err)
	require.Equal(t, "isolated-id", gjson.GetBytes(body, "prompt_cache_key").String())
	require.False(t, gjson.GetBytes(body, "tools").Exists())
	require.False(t, gjson.GetBytes(body, "tool_choice").Exists())
}

func TestV0152PatchGrokResponsesBodySanitizesComposerReasoning(t *testing.T) {
	body := []byte(`{"model":"grok","input":"hello","reasoning":{"effort":"high"},"reasoning_effort":"high","reasoningEffort":"high"}`)

	patched, err := patchGrokResponsesBody(body, "provider/grok-composer-2.5-fast")

	require.NoError(t, err)
	require.False(t, gjson.GetBytes(patched, "reasoning").Exists())
	require.False(t, gjson.GetBytes(patched, "reasoning_effort").Exists())
	require.False(t, gjson.GetBytes(patched, "reasoningEffort").Exists())
}

func TestV0152GrokChatResponsesBridgeEligibility(t *testing.T) {
	eligible, reason := grokChatResponsesBridgeEligibility([]byte(`{"model":"grok","messages":[{"role":"user","content":"hi"}],"stream":false}`))
	require.True(t, eligible)
	require.Empty(t, reason)

	eligible, reason = grokChatResponsesBridgeEligibility([]byte(`{"model":"grok","messages":[{"role":"user","content":"hi"}],"stop":"done"}`))
	require.False(t, eligible)
	require.Equal(t, "unsupported_stop", reason)
}

func TestV0152GrokAPIKeyResponsesRequestUsesOfficialEndpoint(t *testing.T) {
	account := &Account{
		ID:       301,
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "xai-test-key",
		},
	}

	req, err := buildGrokResponsesRequest(context.Background(), nil, account, []byte(`{"model":"grok-4.3"}`), "xai-test-key", "", nil)

	require.NoError(t, err)
	require.Equal(t, "https://api.x.ai/v1/responses", req.URL.String())
	require.Equal(t, "Bearer xai-test-key", req.Header.Get("Authorization"))
}

func TestV0152GrokOAuthOfficialBaseURLMigratesToCLIProxy(t *testing.T) {
	for _, baseURL := range []string{"", "https://api.x.ai", "https://api.x.ai/v1/"} {
		account := &Account{
			Platform: PlatformGrok,
			Type:     AccountTypeOAuth,
			Credentials: map[string]any{
				"base_url": baseURL,
			},
		}
		require.Equal(t, "https://cli-chat-proxy.grok.com/v1", account.GetGrokBaseURL())
	}

	account := &Account{
		Platform: PlatformGrok,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"base_url": "https://grok-proxy.example/v1",
		},
	}
	require.Equal(t, "https://grok-proxy.example/v1", account.GetGrokBaseURL())
}

func TestV0152GrokCatalogFallbackPricing(t *testing.T) {
	svc := &BillingService{cfg: &config.Config{}, fallbackPrices: make(map[string]*ModelPricing)}
	svc.initFallbackPricing()

	for _, model := range []string{"grok-4.20-reasoning", "grok-4.20-non-reasoning"} {
		pricing, err := svc.GetModelPricing(model)
		require.NoError(t, err)
		require.InDelta(t, 1.25e-6, pricing.InputPricePerToken, 1e-12)
		require.InDelta(t, 0.2e-6, pricing.CacheReadPricePerToken, 1e-12)
		require.InDelta(t, 2.5e-6, pricing.OutputPricePerToken, 1e-12)
	}

	for _, model := range []string{"grok-build-0.1", "grok-composer", "composer-2.5"} {
		pricing, err := svc.GetModelPricing(model)
		require.NoError(t, err)
		require.InDelta(t, 1e-6, pricing.InputPricePerToken, 1e-12)
		require.InDelta(t, 0.2e-6, pricing.CacheReadPricePerToken, 1e-12)
		require.InDelta(t, 2e-6, pricing.OutputPricePerToken, 1e-12)
	}
}
