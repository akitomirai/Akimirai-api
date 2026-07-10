package dto

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyFromService_MapsLastUsedAt(t *testing.T) {
	lastUsed := time.Now().UTC().Truncate(time.Second)
	lastUsedIP := "203.0.113.10"
	src := &service.APIKey{
		ID:                 1,
		UserID:             2,
		Key:                "sk-map-last-used",
		Name:               "Mapper",
		Status:             service.StatusActive,
		LastUsedAt:         &lastUsed,
		LastUsedIP:         &lastUsedIP,
		CurrentConcurrency: 3,
	}

	out := APIKeyFromService(src)
	require.NotNil(t, out)
	require.NotNil(t, out.LastUsedAt)
	require.WithinDuration(t, lastUsed, *out.LastUsedAt, time.Second)
	require.NotNil(t, out.LastUsedIP)
	require.Equal(t, lastUsedIP, *out.LastUsedIP)
	require.Equal(t, 3, out.CurrentConcurrency)
}

func TestAPIKeyFromService_MapsNilLastUsedAt(t *testing.T) {
	src := &service.APIKey{
		ID:     1,
		UserID: 2,
		Key:    "sk-map-last-used-nil",
		Name:   "MapperNil",
		Status: service.StatusActive,
	}

	out := APIKeyFromService(src)
	require.NotNil(t, out)
	require.Nil(t, out.LastUsedAt)
	require.Nil(t, out.LastUsedIP)
}

func TestAPIKeyFromService_ReturnsServiceKey(t *testing.T) {
	src := &service.APIKey{
		ID:        1,
		UserID:    2,
		Key:       "ak_hmac_sha256_v1:1111111111111111111111111111111111111111111111111111111111111111",
		KeyHash:   "ak_hmac_sha256_v1:1111111111111111111111111111111111111111111111111111111111111111",
		KeyPrefix: "sk-12345",
		Name:      "MapperMask",
		Status:    service.StatusActive,
	}

	out := APIKeyFromService(src)
	require.NotNil(t, out)
	require.Equal(t, "ak_hmac_sha256_v1:1111111111111111111111111111111111111111111111111111111111111111", out.Key)
	require.Equal(t, "sk-12345", out.KeyPrefix)
	require.False(t, out.KeyVisibleOnce)

	src.Key = "sk-1234567890abcdef"
	src.KeyVisibleOnce = true
	out = APIKeyFromService(src)
	require.NotNil(t, out)
	require.Equal(t, "sk-1234567890abcdef", out.Key)
	require.True(t, out.KeyVisibleOnce)
}
