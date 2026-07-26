package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newOpenAIAccountHeaderTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/1/test", nil)
	return c, rec
}

type accountTestOpenAITokenCacheStub struct {
	tokens map[string]string
}

func (s *accountTestOpenAITokenCacheStub) GetAccessToken(_ context.Context, cacheKey string) (string, error) {
	return s.tokens[cacheKey], nil
}

func (s *accountTestOpenAITokenCacheStub) SetAccessToken(_ context.Context, cacheKey string, token string, _ time.Duration) error {
	if s.tokens == nil {
		s.tokens = map[string]string{}
	}
	s.tokens[cacheKey] = token
	return nil
}

func (s *accountTestOpenAITokenCacheStub) DeleteAccessToken(_ context.Context, cacheKey string) error {
	delete(s.tokens, cacheKey)
	return nil
}

func (s *accountTestOpenAITokenCacheStub) AcquireRefreshLock(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

func (s *accountTestOpenAITokenCacheStub) ReleaseRefreshLock(_ context.Context, _ string) error {
	return nil
}

func TestAccountTestService_OpenAIAPIKeyResponsesUsesAuthorizationAndCodexProbeHeaders(t *testing.T) {
	ctx, recorder := newOpenAIAccountHeaderTestContext()

	upstreamBody := strings.Join([]string{
		`data: {"type":"response.output_text.delta","delta":"pong"}`,
		"",
		`data: {"type":"response.completed"}`,
		"",
	}, "\n")
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(upstreamBody)),
	}}
	svc := &AccountTestService{
		httpUpstream: upstream,
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
	}
	account := &Account{
		ID:          95,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "sk-test",
			"base_url": "https://compat-upstream.example/v1",
		},
		Extra: map[string]any{openai_compat.ExtraKeyResponsesSupported: true},
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.5", "", "")
	require.NoError(t, err)
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, "https://compat-upstream.example/v1/responses", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer sk-test", upstream.lastReq.Header.Get("Authorization"))
	require.Empty(t, upstream.lastReq.Header.Get("Cookie"))
	require.Equal(t, "text/event-stream", upstream.lastReq.Header.Get("Accept"))
	requireOpenAICodexProbeHeaders(t, upstream.lastReq.Header)
	require.Contains(t, recorder.Body.String(), "pong")
	require.Contains(t, recorder.Body.String(), `"success":true`)
}

func TestAccountTestService_OpenAIOAuthResponsesUsesBearerEndUserAuth(t *testing.T) {
	ctx, recorder := newOpenAIAccountHeaderTestContext()

	upstreamBody := strings.Join([]string{
		`data: {"type":"response.output_text.delta","delta":"pong"}`,
		"",
		`data: {"type":"response.completed"}`,
		"",
	}, "\n")
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(upstreamBody)),
	}}
	account := &Account{
		ID:          96,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token":       "oauth-access-token",
			"chatgpt_account_id": "acct_123",
		},
	}
	cache := &accountTestOpenAITokenCacheStub{tokens: map[string]string{
		OpenAITokenCacheKey(account): "oauth-provider-token",
	}}
	svc := &AccountTestService{
		httpUpstream:        upstream,
		openAITokenProvider: NewOpenAITokenProvider(nil, cache, nil),
	}

	err := svc.testOpenAIAccountConnection(ctx, account, "gpt-5.5", "", "")
	require.NoError(t, err)
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, chatgptCodexAPIURL, upstream.lastReq.URL.String())
	require.Equal(t, "chatgpt.com", upstream.lastReq.Host)
	require.Equal(t, "Bearer oauth-provider-token", upstream.lastReq.Header.Get("Authorization"))
	require.Empty(t, upstream.lastReq.Header.Get("Cookie"))
	require.Equal(t, "acct_123", upstream.lastReq.Header.Get("chatgpt-account-id"))
	require.Equal(t, "responses=experimental", upstream.lastReq.Header.Get("OpenAI-Beta"))
	require.Equal(t, "codex_cli_rs", upstream.lastReq.Header.Get("Originator"))
	require.Equal(t, codexCLIVersion, upstream.lastReq.Header.Get("Version"))
	require.Equal(t, codexCLIUserAgent, upstream.lastReq.Header.Get("User-Agent"))
	require.Contains(t, recorder.Body.String(), "pong")
	require.Contains(t, recorder.Body.String(), `"success":true`)
}
