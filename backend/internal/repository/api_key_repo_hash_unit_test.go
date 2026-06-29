package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyRepository_GetByKeyForAuth_FindsHashedKey_SQLite(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "getbykey-auth-hash-unit@test.com")

	rawKey := "sk-hashed-repository-key"
	hash := service.HashAPIKeyWithConfig(rawKey, &config.Config{})
	key := &service.APIKey{
		UserID:    user.ID,
		Key:       "__hashed__sk-hashe__placeholder",
		KeyHash:   hash,
		KeyPrefix: "sk-hashe",
		Name:      "Hashed Key Unit",
		Status:    service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))

	got, err := repo.GetByKeyForAuth(ctx, hash)
	require.NoError(t, err)
	require.Equal(t, key.ID, got.ID)
	require.Equal(t, hash, got.Key)
	require.Equal(t, hash, got.KeyHash)
	require.Equal(t, "sk-hashe", got.KeyPrefix)

	_, err = repo.GetByKeyForAuth(ctx, rawKey)
	require.ErrorIs(t, err, service.ErrAPIKeyNotFound)
}
