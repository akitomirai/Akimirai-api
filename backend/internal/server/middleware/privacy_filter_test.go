//go:build unit

package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/util/privacyfilter"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestEffectivePrivacyFilterConfigMergesGlobalAndUserTypes(t *testing.T) {
	c := privacyFilterContextWithUserConfig(service.PrivacyFilterConfig{
		Enabled: true,
		Types:   []string{privacyfilter.TypeEmail, privacyfilter.TypeToken},
	})

	config := effectivePrivacyFilterConfig(privacyfilter.Config{
		Enabled: true,
		Types:   []string{privacyfilter.TypeAPIKey, privacyfilter.TypeEmail},
	}, c)

	require.True(t, config.Enabled)
	require.Equal(t, []string{privacyfilter.TypeAPIKey, privacyfilter.TypeEmail, privacyfilter.TypeToken}, config.Types)
}

func TestEffectivePrivacyFilterConfigKeepsGlobalWhenUserDisabled(t *testing.T) {
	c := privacyFilterContextWithUserConfig(service.PrivacyFilterConfig{
		Enabled: false,
		Types:   []string{privacyfilter.TypeEmail},
	})

	config := effectivePrivacyFilterConfig(privacyfilter.Config{
		Enabled: true,
		Types:   []string{privacyfilter.TypeAPIKey},
	}, c)

	require.True(t, config.Enabled)
	require.Equal(t, []string{privacyfilter.TypeAPIKey}, config.Types)
}

func TestEffectivePrivacyFilterConfigUsesUserWhenGlobalDisabled(t *testing.T) {
	c := privacyFilterContextWithUserConfig(service.PrivacyFilterConfig{
		Enabled: true,
		Types:   []string{privacyfilter.TypeEmail},
	})

	config := effectivePrivacyFilterConfig(privacyfilter.Config{
		Enabled: false,
		Types:   []string{privacyfilter.TypeAPIKey},
	}, c)

	require.True(t, config.Enabled)
	require.Equal(t, []string{privacyfilter.TypeEmail}, config.Types)
}

func privacyFilterContextWithUserConfig(config service.PrivacyFilterConfig) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(string(ContextKeyAPIKey), &service.APIKey{
		User: &service.User{
			PrivacyFilterConfig: config,
		},
	})
	return c
}
