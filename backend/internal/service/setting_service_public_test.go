//go:build unit

package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type settingPublicRepoStub struct {
	values map[string]string
}

type settingPublicAccountRepoStub struct {
	AccountRepository
	accounts []Account
	err      error
}

type settingPublicPlatformProjectionRepoStub struct {
	AccountRepository
	sources         []ConfiguredAIPlatformSource
	err             error
	projectionCalls atomic.Int64
	fallbackCalls   atomic.Int64
}

func (s *settingPublicPlatformProjectionRepoStub) ListSchedulableAIPlatformSources(context.Context) ([]ConfiguredAIPlatformSource, error) {
	s.projectionCalls.Add(1)
	return s.sources, s.err
}

func (s *settingPublicPlatformProjectionRepoStub) ListSchedulable(context.Context) ([]Account, error) {
	s.fallbackCalls.Add(1)
	return nil, errors.New("full schedulable account hydration must not be used")
}

func (s *settingPublicAccountRepoStub) ListSchedulable(context.Context) ([]Account, error) {
	return s.accounts, s.err
}

func (s *settingPublicRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingPublicRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *settingPublicRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingPublicRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *settingPublicRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *settingPublicRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingPublicRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_GetPublicSettings_ExposesRegistrationEmailSuffixWhitelist(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyRegistrationEnabled:              "true",
			SettingKeyEmailVerifyEnabled:               "true",
			SettingKeyRegistrationEmailSuffixWhitelist: `["@EXAMPLE.com"," @foo.bar ","*.EDU.CN","@invalid_domain",""]`,
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"@example.com", "@foo.bar", "*.edu.cn"}, settings.RegistrationEmailSuffixWhitelist)
}

func TestSettingService_GetPublicSettings_ExposesTablePreferences(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyTableDefaultPageSize: "50",
			SettingKeyTablePageSizeOptions: "[20,50,100]",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, 50, settings.TableDefaultPageSize)
	require.Equal(t, []int{20, 50, 100}, settings.TablePageSizeOptions)
}

func TestSettingService_GetPublicSettings_ExposesForceEmailOnThirdPartySignup(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyForceEmailOnThirdPartySignup: "true",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.ForceEmailOnThirdPartySignup)
}

func TestSettingService_GetPublicSettings_ExposesAllowUserViewErrorRequests(t *testing.T) {
	repo := &settingPublicRepoStub{
		values: map[string]string{
			SettingKeyAllowUserViewErrorRequests: "true",
		},
	}
	svc := NewSettingService(repo, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.AllowUserViewErrorRequests)
}

func TestSettingService_GetPublicSettings_ExposesWeChatOAuthModeCapabilities(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyWeChatConnectEnabled:             "true",
			SettingKeyWeChatConnectAppID:               "wx-mp-app",
			SettingKeyWeChatConnectAppSecret:           "wx-mp-secret",
			SettingKeyWeChatConnectMode:                "mp",
			SettingKeyWeChatConnectScopes:              "snsapi_base",
			SettingKeyWeChatConnectOpenEnabled:         "true",
			SettingKeyWeChatConnectMPEnabled:           "true",
			SettingKeyWeChatConnectRedirectURL:         "https://api.example.com/api/v1/auth/oauth/wechat/callback",
			SettingKeyWeChatConnectFrontendRedirectURL: "/auth/wechat/callback",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.WeChatOAuthEnabled)
	require.True(t, settings.WeChatOAuthOpenEnabled)
	require.True(t, settings.WeChatOAuthMPEnabled)
}

func TestSettingService_GetPublicSettings_DoesNotExposeMobileOnlyWeChatAsWebOAuthAvailable(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{
		values: map[string]string{
			SettingKeyWeChatConnectEnabled:             "true",
			SettingKeyWeChatConnectMobileEnabled:       "true",
			SettingKeyWeChatConnectMode:                "mobile",
			SettingKeyWeChatConnectMobileAppID:         "wx-mobile-app",
			SettingKeyWeChatConnectMobileAppSecret:     "wx-mobile-secret",
			SettingKeyWeChatConnectFrontendRedirectURL: "/auth/wechat/callback",
		},
	}, &config.Config{})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.False(t, settings.WeChatOAuthEnabled)
	require.False(t, settings.WeChatOAuthOpenEnabled)
	require.False(t, settings.WeChatOAuthMPEnabled)
	require.True(t, settings.WeChatOAuthMobileEnabled)
}

func TestSettingService_GetPublicSettings_FallsBackToConfigForWeChatOAuthCapabilities(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{
		WeChat: config.WeChatConnectConfig{
			Enabled:             true,
			OpenEnabled:         true,
			OpenAppID:           "wx-open-config",
			OpenAppSecret:       "wx-open-secret",
			FrontendRedirectURL: "/auth/wechat/config-callback",
		},
	})

	settings, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	require.True(t, settings.WeChatOAuthEnabled)
	require.True(t, settings.WeChatOAuthOpenEnabled)
	require.False(t, settings.WeChatOAuthMPEnabled)
	require.False(t, settings.WeChatOAuthMobileEnabled)
}

func TestConfiguredAIPlatformLabelsFromAccounts_DerivesFamiliesFromPlatformAndMappings(t *testing.T) {
	labels := configuredAIPlatformLabelsFromAccounts([]Account{
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"glm":      "glm-4.6",
					"deepseek": "deepseek-chat",
				},
			},
		},
		{Platform: PlatformAnthropic},
		{Platform: PlatformGrok},
	})

	require.Equal(t, []string{"GPT", "Claude", "GLM", "DeepSeek", "Grok"}, labels)
}

func TestSettingService_GetPublicSettings_ExposesConfiguredAIPlatforms(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})
	svc.SetAccountRepository(&settingPublicAccountRepoStub{accounts: []Account{
		{Platform: PlatformGemini},
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"claude": "claude-sonnet-4-5"},
			},
		},
	}})

	settings, err := svc.GetPublicSettings(context.Background())

	require.NoError(t, err)
	require.Equal(t, []string{"GPT", "Claude", "Gemini"}, settings.ConfiguredAIPlatforms)
}

func TestSettingService_GetPublicSettings_FailsClosedWhenConfiguredPlatformsUnavailable(t *testing.T) {
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})
	svc.SetAccountRepository(&settingPublicAccountRepoStub{err: errors.New("database unavailable")})

	settings, err := svc.GetPublicSettings(context.Background())

	require.NoError(t, err)
	require.Empty(t, settings.ConfiguredAIPlatforms)
}

func TestSettingService_GetPublicSettings_UsesCachedPlatformProjection(t *testing.T) {
	repo := &settingPublicPlatformProjectionRepoStub{sources: []ConfiguredAIPlatformSource{
		{Platform: PlatformGemini},
		{
			Platform:     PlatformOpenAI,
			ModelMapping: map[string]string{"claude": "claude-sonnet-4-5"},
		},
	}}
	svc := NewSettingService(&settingPublicRepoStub{values: map[string]string{}}, &config.Config{})
	svc.SetAccountRepository(repo)

	first, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)
	second, err := svc.GetPublicSettings(context.Background())
	require.NoError(t, err)

	require.Equal(t, []string{"GPT", "Claude", "Gemini"}, first.ConfiguredAIPlatforms)
	require.Equal(t, first.ConfiguredAIPlatforms, second.ConfiguredAIPlatforms)
	require.Equal(t, int64(1), repo.projectionCalls.Load(), "the short-lived cache must collapse repeated public requests")
	require.Zero(t, repo.fallbackCalls.Load(), "the lightweight projection must own this public capability lookup")
}
