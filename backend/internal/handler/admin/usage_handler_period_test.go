package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseOptionalAdminUsageTimeRangeSupportsRollingPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/?period=24h&timezone=UTC", nil)

	startTime, endTime, ok := parseOptionalAdminUsageTimeRange(context)
	require.True(t, ok)
	require.NotNil(t, startTime)
	require.NotNil(t, endTime)
	require.Equal(t, 24*time.Hour, endTime.Sub(*startTime))
}

func TestParseAdminUsageTimeRangeRejectsInvalidPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/?period=invalid", nil)

	_, _, ok := parseOptionalAdminUsageTimeRange(context)
	require.False(t, ok)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestParseAdminUsageStatsTimeRangeSupportsDashboardPeriods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodGet, "/?period=48h&timezone=UTC", nil)

	startTime, endTime, ok := parseAdminUsageStatsTimeRange(context)
	require.True(t, ok)
	require.Equal(t, 48*time.Hour, endTime.Sub(startTime))
}
