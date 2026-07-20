package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNormalizeDashboardGranularity(t *testing.T) {
	for _, value := range []string{"hour", "2h", "4h", "8h", "day"} {
		require.Equal(t, value, normalizeDashboardGranularity(value))
	}
	require.Equal(t, "day", normalizeDashboardGranularity("minute"))
	require.Equal(t, "day", normalizeDashboardGranularity("2h; DROP TABLE usage_logs"))
}

func TestResolveDashboardPeriodRange(t *testing.T) {
	now := time.Date(2026, 7, 13, 15, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	start, end, ok := resolveDashboardPeriodRange(now, "Asia/Shanghai", "48h")
	require.True(t, ok)
	require.Equal(t, now.Add(-48*time.Hour), start)
	require.Equal(t, now, end)

	_, _, ok = resolveDashboardPeriodRange(now, "Asia/Shanghai", "48h; DROP TABLE usage_logs")
	require.False(t, ok)
}

func TestDashboardResponseEndDateUsesRollingPeriodEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	end := time.Date(2026, 7, 13, 15, 30, 0, 0, time.FixedZone("CST", 8*60*60))

	periodContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	periodContext.Request = httptest.NewRequest(http.MethodGet, "/?period=24h", nil)
	require.Equal(t, "2026-07-13", dashboardResponseEndDate(periodContext, end))

	yesterdayContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	yesterdayContext.Request = httptest.NewRequest(http.MethodGet, "/?period=yesterday", nil)
	require.Equal(t, "2026-07-12", dashboardResponseEndDate(yesterdayContext, end))

	calendarContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	calendarContext.Request = httptest.NewRequest(http.MethodGet, "/?end_date=2026-07-13", nil)
	require.Equal(t, "2026-07-12", dashboardResponseEndDate(calendarContext, end))
}

func TestParseTimeRangeUsesNextLocalMidnightAcrossDST(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, testCase := range []struct {
		name     string
		date     string
		duration time.Duration
	}{
		{name: "spring forward", date: "2026-03-08", duration: 23 * time.Hour},
		{name: "fall back", date: "2026-11-01", duration: 25 * time.Hour},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest(
				http.MethodGet,
				"/?start_date="+testCase.date+"&end_date="+testCase.date+"&timezone=America%2FNew_York",
				nil,
			)
			start, end, err := parseTimeRange(context)
			require.NoError(t, err)
			require.Equal(t, testCase.duration, end.Sub(start))
		})
	}
}

func TestDashboardResponseCacheMetadataOwnsTimezonePeriodAndDates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utc := time.UTC
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	startInstant := time.Date(2026, 7, 12, 16, 0, 0, 0, utc)
	endInstant := startInstant.Add(24 * time.Hour)

	utcContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	utcContext.Request = httptest.NewRequest(http.MethodGet, "/?period=24h&timezone=UTC", nil)
	utcMetadata := newDashboardResponseCacheMetadata(utcContext, startInstant, endInstant)

	shanghaiContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	shanghaiContext.Request = httptest.NewRequest(http.MethodGet, "/?period=today&timezone=Asia%2FShanghai", nil)
	shanghaiMetadata := newDashboardResponseCacheMetadata(
		shanghaiContext,
		startInstant.In(shanghai),
		endInstant.In(shanghai),
	)

	require.True(t, startInstant.Equal(startInstant.In(shanghai)))
	require.NotEqual(t, utcMetadata, shanghaiMetadata)
	require.NotEqual(t,
		mustMarshalDashboardCacheKey(dashboardSnapshotV2CacheKey{
			StartTime: startInstant.Format(time.RFC3339),
			EndTime:   endInstant.Format(time.RFC3339),
			Response:  utcMetadata,
		}),
		mustMarshalDashboardCacheKey(dashboardSnapshotV2CacheKey{
			StartTime: startInstant.Format(time.RFC3339),
			EndTime:   endInstant.Format(time.RFC3339),
			Response:  shanghaiMetadata,
		}),
	)
	require.NotEqual(t,
		mustMarshalDashboardCacheKey(dashboardUsersRankingCacheKey{
			StartTime: startInstant.Format(time.RFC3339),
			EndTime:   endInstant.Format(time.RFC3339),
			Response:  utcMetadata,
			Limit:     12,
		}),
		mustMarshalDashboardCacheKey(dashboardUsersRankingCacheKey{
			StartTime: startInstant.Format(time.RFC3339),
			EndTime:   endInstant.Format(time.RFC3339),
			Response:  shanghaiMetadata,
			Limit:     12,
		}),
	)
}
