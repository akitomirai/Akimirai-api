package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type grokNoAccountDiagnoser struct {
	platform string
}

func (d *grokNoAccountDiagnoser) DiagnoseModelAvailabilityForPlatform(
	_ context.Context,
	_ *int64,
	_ string,
	platform string,
) service.ModelAvailabilityDiagnosis {
	d.platform = platform
	return service.ModelAvailabilityDiagnosis{HasAccountsInPool: true, HasModelSupport: false}
}

func TestV0152OpenAICompatibleNoAccountDiagnosisUsesGrokPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	groupID := int64(43)
	apiKey := &service.APIKey{
		GroupID: &groupID,
		Group:   &service.Group{ID: groupID, Platform: service.PlatformGrok},
	}
	diagnoser := &grokNoAccountDiagnoser{}

	classification := classifyOpenAICompatibleNoAccountErrorFromGin(c, diagnoser, apiKey, "grok-4.5", "grok-4.5")

	require.Equal(t, http.StatusNotFound, classification.Status)
	require.Equal(t, "model_not_found", classification.ErrType)
	require.True(t, classification.ModelNotFound)
	require.Equal(t, service.PlatformGrok, diagnoser.platform)
	require.EqualError(t,
		openAICompatibleSelectionErrorForLog(fmt.Errorf("no available OpenAI accounts supporting model: grok-4.5"), service.PlatformGrok),
		"no available Grok accounts supporting model: grok-4.5",
	)
}
