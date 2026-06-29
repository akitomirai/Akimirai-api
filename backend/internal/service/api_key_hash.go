package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

const apiKeyHashPrefix = "ak_hmac_sha256_v1:"

func (s *APIKeyService) hashAPIKey(raw string) string {
	if s == nil {
		return HashAPIKeyWithConfig(raw, nil)
	}
	return HashAPIKeyWithConfig(raw, s.cfg)
}

func HashAPIKeyWithConfig(raw string, cfg *config.Config) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(apiKeyHashSecretFromConfig(cfg)))
	_, _ = mac.Write([]byte(raw))
	return apiKeyHashPrefix + hex.EncodeToString(mac.Sum(nil))
}

func (s *APIKeyService) apiKeyHashSecret() string {
	if s == nil {
		return apiKeyHashSecretFromConfig(nil)
	}
	return apiKeyHashSecretFromConfig(s.cfg)
}

func apiKeyHashSecretFromConfig(cfg *config.Config) string {
	if cfg != nil {
		if secret := strings.TrimSpace(cfg.JWT.Secret); secret != "" {
			return secret
		}
		if secret := strings.TrimSpace(cfg.Totp.EncryptionKey); secret != "" {
			return secret
		}
	}
	return "sub2api-api-key-hash-development-fallback"
}

func IsAPIKeyHash(value string) bool {
	_, ok := apiKeyHashCacheKey(value)
	return ok
}

func apiKeyHashCacheKey(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, apiKeyHashPrefix) {
		return "", false
	}
	digest := strings.TrimPrefix(value, apiKeyHashPrefix)
	if len(digest) != sha256.Size*2 {
		return "", false
	}
	for _, ch := range digest {
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') {
			continue
		}
		return "", false
	}
	return digest, true
}

func apiKeyDisplayPrefix(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) <= 8 {
		return raw
	}
	return raw[:8]
}

func apiKeyStoragePlaceholder(hash, prefix string) string {
	hash = strings.TrimSpace(hash)
	prefix = strings.TrimSpace(prefix)
	if digest, ok := apiKeyHashCacheKey(hash); ok {
		if len(digest) > 24 {
			digest = digest[:24]
		}
		return "__hashed__" + prefix + "__" + digest
	}
	return "__hashed__" + prefix
}

func MaskAPIKey(prefix, raw string) string {
	prefix = strings.TrimSpace(prefix)
	raw = strings.TrimSpace(raw)
	if prefix == "" {
		prefix = apiKeyDisplayPrefix(raw)
	}
	if prefix == "" {
		return ""
	}
	if strings.HasPrefix(raw, "__hashed__") || strings.HasPrefix(raw, apiKeyHashPrefix) {
		return prefix + "..."
	}
	if len(raw) <= len(prefix)+4 {
		return prefix + "..."
	}
	return prefix + "..." + raw[len(raw)-4:]
}
