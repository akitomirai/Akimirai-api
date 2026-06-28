package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/util/privacyfilter"
	"github.com/gin-gonic/gin"
)

func PrivacyFilter(settingService *service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if settingService == nil || !methodMayHaveBody(c.Request.Method) || !isJSONContentType(c.GetHeader("Content-Type")) {
			c.Next()
			return
		}

		config := settingService.GetPrivacyFilterConfig(c.Request.Context())
		if !config.Enabled || len(config.Types) == 0 {
			c.Next()
			return
		}

		if c.Request.Body == nil {
			c.Next()
			return
		}

		originalBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			_ = c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewReader(originalBody))
			slog.Warn("privacy filter: failed to read request body", "error", err)
			c.Next()
			return
		}
		_ = c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewReader(originalBody))

		filteredBody, changed, err := privacyfilter.RedactJSONBody(originalBody, config)
		if err != nil {
			slog.Warn("privacy filter: failed to redact request body", "error", err)
			c.Next()
			return
		}
		if !changed {
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(filteredBody))
		c.Request.ContentLength = int64(len(filteredBody))
		c.Request.Header.Set("Content-Length", strconv.Itoa(len(filteredBody)))
		c.Next()
	}
}

func methodMayHaveBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

func isJSONContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/json")
}
