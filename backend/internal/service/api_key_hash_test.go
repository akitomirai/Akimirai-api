package service

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestHashAPIKeyWithConfig_UsesVersionedHMAC(t *testing.T) {
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-secret"

	hash := HashAPIKeyWithConfig("sk-test-secret", cfg)

	require.True(t, IsAPIKeyHash(hash))
	require.True(t, strings.HasPrefix(hash, apiKeyHashPrefix))
	require.NotContains(t, hash, "sk-test-secret")
	require.Equal(t, hash, HashAPIKeyWithConfig(" sk-test-secret ", cfg))
	require.NotEqual(t, hash, HashAPIKeyWithConfig("sk-other-secret", cfg))
}

func TestAPIKeyDisplayKey_MasksStoredHashButShowsOneTimeSecret(t *testing.T) {
	raw := "sk-1234567890abcdef"
	hash := HashAPIKeyWithConfig(raw, &config.Config{})
	key := &APIKey{
		Key:       hash,
		KeyHash:   hash,
		KeyPrefix: apiKeyDisplayPrefix(raw),
	}

	require.Equal(t, "sk-12345...", key.DisplayKey())

	key.Key = raw
	key.KeyVisibleOnce = true
	require.Equal(t, raw, key.DisplayKey())
}
