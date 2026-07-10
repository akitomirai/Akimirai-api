package service

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
)

const (
	defaultOpenAICompatibleUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36 Edg/149.0.0.0"
	defaultGrokCompatibleUserAgent   = "sub2api-grok/1.0"
)

func openAICompatibleUserAgent(account *Account) string {
	if account != nil {
		if customUA := strings.TrimSpace(account.GetOpenAIUserAgent()); customUA != "" {
			return customUA
		}
		if account.Platform == PlatformGrok {
			return defaultGrokCompatibleUserAgent
		}
	}
	return defaultOpenAICompatibleUserAgent
}

func applyOpenAICompatibleUserAgent(req *http.Request, account *Account) {
	if req == nil {
		return
	}
	req.Header.Set("User-Agent", openAICompatibleUserAgent(account))
}

func applyOpenAIAPIKeyCompatibleUserAgent(req *http.Request, account *Account) {
	if account == nil || account.Type != AccountTypeAPIKey {
		return
	}
	applyOpenAICompatibleUserAgent(req, account)
}

func logOpenAIUpstreamHTTPFailure(scope string, account *Account, upstreamURL string, model string, statusCode int, body []byte) {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = "service.openai_upstream"
	}
	accountID := int64(0)
	if account != nil {
		accountID = account.ID
	}
	logger.LegacyPrintf(scope,
		"openai_upstream_http_failed: account_id=%d upstream_url=%s model=%s status=%d code=%s",
		accountID,
		strings.TrimSpace(upstreamURL),
		strings.TrimSpace(model),
		statusCode,
		extractOpenAIUpstreamErrorCodeForLog(body),
	)
}

func extractOpenAIUpstreamErrorCodeForLog(body []byte) string {
	for _, path := range []string{"code", "detail.code", "error.code", "error.type", "type"} {
		if code := strings.TrimSpace(gjson.GetBytes(body, path).String()); code != "" {
			return code
		}
	}
	return ""
}
