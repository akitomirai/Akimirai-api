package repository

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestOpenAIWSIngressLeaseEnforcesPerAPIKeyLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := NewConcurrencyCache(rdb, 15, 60).(*concurrencyCache)
	ctx := context.Background()

	acquired, err := cache.AcquireOpenAIWSIngressLease(ctx, 42, 1, "lease-a")
	require.NoError(t, err)
	require.True(t, acquired)

	acquired, err = cache.AcquireOpenAIWSIngressLease(ctx, 42, 1, "lease-b")
	require.NoError(t, err)
	require.False(t, acquired)

	owned, err := cache.RefreshOpenAIWSIngressLease(ctx, 42, "lease-a")
	require.NoError(t, err)
	require.True(t, owned)
	require.NoError(t, cache.ReleaseOpenAIWSIngressLease(ctx, 42, "lease-a"))

	acquired, err = cache.AcquireOpenAIWSIngressLease(ctx, 42, 1, "lease-b")
	require.NoError(t, err)
	require.True(t, acquired)
}
