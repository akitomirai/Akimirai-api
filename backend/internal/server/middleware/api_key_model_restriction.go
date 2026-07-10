package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/googleapi"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// APIKeyModelRestriction enforces API-key model allowlists after authentication.
// JSON bodies are restored so downstream handlers retain their existing parsing behavior.
func APIKeyModelRestriction() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := GetAPIKeyFromContext(c)
		if !ok || apiKey == nil || len(apiKey.AllowedModels) == 0 {
			c.Next()
			return
		}

		model := apiKeyModelFromRequest(c)
		if model == "" || !RejectAPIKeyModel(c, apiKey, model) {
			c.Next()
			return
		}
	}
}

// RejectAPIKeyModel writes the protocol-compatible rejection for a parsed model.
// It returns true only when the allowlist rejects the model.
func RejectAPIKeyModel(c *gin.Context, apiKey *service.APIKey, model string) bool {
	if apiKey == nil || apiKey.AllowsModel(model) {
		return false
	}
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
	writeAPIKeyModelNotAllowed(c, model)
	c.Abort()
	return true
}

func apiKeyModelFromRequest(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if model := geminiModelFromPath(c); model != "" {
		return model
	}
	if c.Request.Method == http.MethodGet || c.Request.Body == nil {
		return ""
	}
	contentType := strings.ToLower(c.ContentType())
	if strings.Contains(contentType, "multipart/form-data") {
		return ""
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if !gjson.ValidBytes(body) {
		return ""
	}
	result := gjson.GetBytes(body, "model")
	if !result.Exists() || result.Type != gjson.String {
		return ""
	}
	return strings.TrimSpace(result.String())
}

func geminiModelFromPath(c *gin.Context) string {
	for _, param := range []string{"model", "modelAction"} {
		value := strings.TrimSpace(strings.TrimPrefix(c.Param(param), "/"))
		value = strings.TrimPrefix(value, "models/")
		if idx := strings.IndexByte(value, ':'); idx >= 0 {
			value = value[:idx]
		}
		if value != "" {
			return value
		}
	}
	return ""
}

func writeAPIKeyModelNotAllowed(c *gin.Context, model string) {
	message := service.APIKeyModelNotAllowedMessage(model)
	path := strings.ToLower(c.Request.URL.Path)
	if strings.Contains(path, "/v1beta/") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    http.StatusForbidden,
				"message": message,
				"status":  googleapi.HTTPStatusToGoogleStatus(http.StatusForbidden),
				"details": []gin.H{{"reason": service.APIKeyModelNotAllowedCode}},
			},
		})
		return
	}
	if strings.HasSuffix(path, "/messages") || strings.HasSuffix(path, "/messages/count_tokens") {
		c.JSON(http.StatusForbidden, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "permission_error",
				"code":    service.APIKeyModelNotAllowedCode,
				"message": message,
			},
		})
		return
	}
	c.JSON(http.StatusForbidden, gin.H{
		"error": gin.H{
			"type":    "permission_error",
			"code":    service.APIKeyModelNotAllowedCode,
			"message": message,
		},
	})
}
