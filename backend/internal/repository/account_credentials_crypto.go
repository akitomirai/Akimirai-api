package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const accountCredentialCipherPrefix = "sub2api_enc_v1:"

func (r *accountRepository) credentialsForStorage(credentials map[string]any) (map[string]any, error) {
	return accountCredentialsForStorage(credentials, r.credentialEncryptor)
}

func (r *accountRepository) credentialsForService(credentials map[string]any) (map[string]any, error) {
	return accountCredentialsForService(credentials, r.credentialEncryptor)
}

func (r *accountRepository) accountEntityToService(m *dbent.Account) (*service.Account, error) {
	out := accountEntityToService(m)
	if out == nil {
		return nil, nil
	}

	credentials, err := r.credentialsForService(out.Credentials)
	if err != nil {
		return nil, err
	}
	out.Credentials = credentials
	return out, nil
}

func accountCredentialsForStorage(credentials map[string]any, encryptor service.SecretEncryptor) (map[string]any, error) {
	if credentials == nil {
		return map[string]any{}, nil
	}

	out := make(map[string]any, len(credentials))
	for key, value := range credentials {
		converted, err := accountCredentialValueForStorage(key, value, encryptor)
		if err != nil {
			return nil, err
		}
		out[key] = converted
	}
	return out, nil
}

func accountCredentialValueForStorage(key string, value any, encryptor service.SecretEncryptor) (any, error) {
	if isSensitiveAccountCredentialKey(key) {
		return encryptAccountCredentialValue(key, value, encryptor)
	}

	switch typed := value.(type) {
	case map[string]any:
		return accountCredentialsForStorage(typed, encryptor)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			converted, err := accountCredentialValueForStorage("", item, encryptor)
			if err != nil {
				return nil, err
			}
			out[i] = converted
		}
		return out, nil
	default:
		return value, nil
	}
}

func encryptAccountCredentialValue(key string, value any, encryptor service.SecretEncryptor) (any, error) {
	if value == nil || encryptor == nil {
		return value, nil
	}
	if s, ok := value.(string); ok {
		if strings.TrimSpace(s) == "" || isEncryptedAccountCredentialValue(s) {
			return s, nil
		}
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal account credential %q for encryption: %w", key, err)
	}
	ciphertext, err := encryptor.Encrypt(string(payload))
	if err != nil {
		return nil, fmt.Errorf("encrypt account credential %q: %w", key, err)
	}
	return accountCredentialCipherPrefix + ciphertext, nil
}

func accountCredentialsForService(credentials map[string]any, encryptor service.SecretEncryptor) (map[string]any, error) {
	if credentials == nil {
		return nil, nil
	}

	out := make(map[string]any, len(credentials))
	for key, value := range credentials {
		converted, err := accountCredentialValueForService(key, value, encryptor)
		if err != nil {
			return nil, err
		}
		out[key] = converted
	}
	return out, nil
}

func accountCredentialValueForService(key string, value any, encryptor service.SecretEncryptor) (any, error) {
	if s, ok := value.(string); ok && isEncryptedAccountCredentialValue(s) {
		if encryptor == nil {
			return nil, fmt.Errorf("decrypt account credential %q: encryptor not configured", key)
		}
		plaintext, err := encryptor.Decrypt(strings.TrimPrefix(s, accountCredentialCipherPrefix))
		if err != nil {
			return nil, fmt.Errorf("decrypt account credential %q: %w", key, err)
		}

		var decoded any
		if err := json.Unmarshal([]byte(plaintext), &decoded); err != nil {
			return plaintext, nil
		}
		return decoded, nil
	}

	switch typed := value.(type) {
	case map[string]any:
		return accountCredentialsForService(typed, encryptor)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			converted, err := accountCredentialValueForService(key, item, encryptor)
			if err != nil {
				return nil, err
			}
			out[i] = converted
		}
		return out, nil
	default:
		return value, nil
	}
}

func isSensitiveAccountCredentialKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	return service.IsSensitiveCredentialKey(key)
}

func isEncryptedAccountCredentialValue(value string) bool {
	return strings.HasPrefix(value, accountCredentialCipherPrefix)
}
