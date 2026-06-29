package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func newAccountRepoEncryptedSQLite(t *testing.T) (*accountRepository, *dbent.Client) {
	t.Helper()

	db, err := sql.Open("sqlite", "file:account_repo_encrypted?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	return newAccountRepositoryWithSQL(client, db, nil, newTestAccountEncryptor(t)), client
}

func newTestAccountEncryptor(t *testing.T) service.SecretEncryptor {
	t.Helper()
	enc, err := NewAESEncryptor(&config.Config{
		Totp: config.TotpConfig{EncryptionKey: strings.Repeat("42", 32)},
	})
	require.NoError(t, err)
	return enc
}

func mustCreatePlainAccountUnit(t *testing.T, ctx context.Context, client *dbent.Client, account *service.Account) *service.Account {
	t.Helper()
	if account.Credentials == nil {
		account.Credentials = map[string]any{}
	}
	if account.Extra == nil {
		account.Extra = map[string]any{}
	}
	if account.Concurrency == 0 {
		account.Concurrency = 1
	}
	if account.Priority == 0 {
		account.Priority = 1
	}

	created, err := client.Account.Create().
		SetName(account.Name).
		SetPlatform(account.Platform).
		SetType(account.Type).
		SetStatus(account.Status).
		SetCredentials(account.Credentials).
		SetExtra(account.Extra).
		SetConcurrency(account.Concurrency).
		SetPriority(account.Priority).
		SetSchedulable(account.Schedulable).
		Save(ctx)
	require.NoError(t, err)

	account.ID = created.ID
	account.CreatedAt = created.CreatedAt
	account.UpdatedAt = created.UpdatedAt
	return account
}

func TestAccountRepository_CreateEncryptsSensitiveCredentials(t *testing.T) {
	repo, client := newAccountRepoEncryptedSQLite(t)
	ctx := context.Background()

	account := &service.Account{
		Name:     "encrypted-account",
		Platform: service.PlatformAnthropic,
		Type:     service.AccountTypeOAuth,
		Status:   service.StatusActive,
		Credentials: map[string]any{
			"access_token":  "sk-access-secret",
			"refresh_token": "sk-refresh-secret",
			"api_key":       "sk-upstream-secret",
			"cookie":        "session=secret-cookie",
			"private_key":   "-----BEGIN TOP LEVEL PRIVATE KEY-----",
			"base_url":      "https://example.invalid",
			"service_account_json": map[string]any{
				"client_email": "bot@example.invalid",
				"private_key":  "-----BEGIN PRIVATE KEY-----",
			},
		},
		Extra: map[string]any{},
	}

	require.NoError(t, repo.Create(ctx, account))

	raw, err := client.Account.Get(ctx, account.ID)
	require.NoError(t, err)

	accessToken, ok := raw.Credentials["access_token"].(string)
	require.True(t, ok)
	require.True(t, strings.HasPrefix(accessToken, accountCredentialCipherPrefix))
	require.NotContains(t, accessToken, "sk-access-secret")

	apiKey, ok := raw.Credentials["api_key"].(string)
	require.True(t, ok)
	require.True(t, strings.HasPrefix(apiKey, accountCredentialCipherPrefix))
	require.NotContains(t, apiKey, "sk-upstream-secret")

	cookie, ok := raw.Credentials["cookie"].(string)
	require.True(t, ok)
	require.True(t, strings.HasPrefix(cookie, accountCredentialCipherPrefix))
	require.NotContains(t, cookie, "secret-cookie")

	privateKey, ok := raw.Credentials["private_key"].(string)
	require.True(t, ok)
	require.True(t, strings.HasPrefix(privateKey, accountCredentialCipherPrefix))
	require.NotContains(t, privateKey, "TOP LEVEL")

	serviceAccount, ok := raw.Credentials["service_account_json"].(string)
	require.True(t, ok)
	require.True(t, strings.HasPrefix(serviceAccount, accountCredentialCipherPrefix))

	got, err := repo.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Equal(t, "sk-access-secret", got.Credentials["access_token"])
	require.Equal(t, "sk-refresh-secret", got.Credentials["refresh_token"])
	require.Equal(t, "sk-upstream-secret", got.Credentials["api_key"])
	require.Equal(t, "session=secret-cookie", got.Credentials["cookie"])
	require.Equal(t, "-----BEGIN TOP LEVEL PRIVATE KEY-----", got.Credentials["private_key"])
	require.Equal(t, "https://example.invalid", got.Credentials["base_url"])

	serviceAccountMap, ok := got.Credentials["service_account_json"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "bot@example.invalid", serviceAccountMap["client_email"])
	require.Equal(t, "-----BEGIN PRIVATE KEY-----", serviceAccountMap["private_key"])
}

func TestAccountRepository_UpdateCredentialsEncryptsAndPreservesPlaintextCompatibility(t *testing.T) {
	repo, client := newAccountRepoEncryptedSQLite(t)
	ctx := context.Background()

	plainAccount := mustCreatePlainAccountUnit(t, ctx, client, &service.Account{
		Name:     "legacy-plain",
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeOAuth,
		Status:   service.StatusActive,
		Credentials: map[string]any{
			"access_token":  "legacy-access-token",
			"base_url":      "https://legacy.example.invalid",
			"refresh_token": "legacy-refresh-token",
		},
		Extra: map[string]any{},
	})

	got, err := repo.GetByID(ctx, plainAccount.ID)
	require.NoError(t, err)
	require.Equal(t, "legacy-access-token", got.Credentials["access_token"])
	require.Equal(t, "legacy-refresh-token", got.Credentials["refresh_token"])

	newCreds := map[string]any{
		"access_token":  "new-access-token",
		"refresh_token": "new-refresh-token",
		"region":        "us",
	}
	require.NoError(t, repo.UpdateCredentials(ctx, plainAccount.ID, newCreds))

	raw, err := client.Account.Get(ctx, plainAccount.ID)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(raw.Credentials["access_token"].(string), accountCredentialCipherPrefix))

	got, err = repo.GetByID(ctx, plainAccount.ID)
	require.NoError(t, err)
	require.Equal(t, "new-access-token", got.Credentials["access_token"])
	require.Equal(t, "new-refresh-token", got.Credentials["refresh_token"])
	require.Equal(t, "us", got.Credentials["region"])
}

func TestAccountCredentialStorageRoundTripKeepsJSONShape(t *testing.T) {
	enc := newTestAccountEncryptor(t)

	stored, err := accountCredentialsForStorage(map[string]any{
		"service_account_json": map[string]any{"private_key": "k", "client_email": "a"},
		"access_token":         "abc",
		"region":               "us",
	}, enc)
	require.NoError(t, err)

	raw, err := json.Marshal(stored)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "abc")
	require.NotContains(t, string(raw), "private_key\":\"k")

	got, err := accountCredentialsForService(stored, enc)
	require.NoError(t, err)
	require.Equal(t, "abc", got["access_token"])
	serviceAccount, ok := got["service_account_json"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "k", serviceAccount["private_key"])
	require.Equal(t, "a", serviceAccount["client_email"])
}
