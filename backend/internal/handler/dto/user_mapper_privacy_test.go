package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/util/privacyfilter"
	"github.com/stretchr/testify/require"
)

func TestUserFromServiceShallowNormalizesPrivacyFilterConfig(t *testing.T) {
	out := UserFromServiceShallow(&service.User{
		ID: 1,
		PrivacyFilterConfig: service.PrivacyFilterConfig{
			Enabled: true,
			Types: []string{
				privacyfilter.TypeToken,
				privacyfilter.TypeEmail,
				"not-real",
				privacyfilter.TypeToken,
			},
		},
	})

	require.NotNil(t, out)
	require.Equal(t, service.PrivacyFilterConfig{
		Enabled: true,
		Types:   []string{privacyfilter.TypeToken, privacyfilter.TypeEmail},
	}, out.PrivacyFilterConfig)
}
