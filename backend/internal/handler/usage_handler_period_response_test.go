package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUserUsageResponseEndDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	end := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)

	yesterdayContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	yesterdayContext.Request = httptest.NewRequest(http.MethodGet, "/?period=yesterday", nil)
	require.Equal(t, "2026-07-12", userUsageResponseEndDate(yesterdayContext, end))

	rollingContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	rollingContext.Request = httptest.NewRequest(http.MethodGet, "/?period=24h", nil)
	require.Equal(t, "2026-07-13", userUsageResponseEndDate(rollingContext, end))

	calendarContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	calendarContext.Request = httptest.NewRequest(http.MethodGet, "/?end_date=2026-07-12", nil)
	require.Equal(t, "2026-07-12", userUsageResponseEndDate(calendarContext, end))

	customContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	customContext.Request = httptest.NewRequest(http.MethodGet, "/?end_time=2026-07-13T22:15", nil)
	require.Equal(t, "2026-07-13", userUsageResponseEndDate(customContext, end))
}
