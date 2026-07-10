package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyModelRestrictionRejectsDeniedJSONModelAndRestoresAllowedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name       string
		model      string
		wantStatus int
		wantNext   bool
	}{
		{name: "allowed", model: "gpt-5", wantStatus: http.StatusNoContent, wantNext: true},
		{name: "denied", model: "gpt-4.1", wantStatus: http.StatusForbidden, wantNext: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set(string(ContextKeyAPIKey), &service.APIKey{AllowedModels: []string{"gpt-5"}})
				c.Next()
			})
			router.Use(APIKeyModelRestriction())
			router.POST("/v1/responses", func(c *gin.Context) {
				body, err := io.ReadAll(c.Request.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), tc.model)
				c.Status(http.StatusNoContent)
			})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"`+tc.model+`"}`))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)

			require.Equal(t, tc.wantStatus, recorder.Code)
			if !tc.wantNext {
				require.Contains(t, recorder.Body.String(), service.APIKeyModelNotAllowedCode)
			}
		})
	}
}

func TestAPIKeyModelRestrictionRejectsDeniedGeminiPathModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyAPIKey), &service.APIKey{AllowedModels: []string{"gemini-2.5-pro"}})
		c.Next()
	})
	router.Use(APIKeyModelRestriction())
	router.POST("/v1beta/models/*modelAction", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-2.0-flash:generateContent", strings.NewReader(`{"contents":[]}`))
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), service.APIKeyModelNotAllowedCode)
}
