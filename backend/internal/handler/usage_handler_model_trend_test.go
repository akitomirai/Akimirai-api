package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type modelTrendUsageRepoStub struct {
	service.UsageLogRepository
	points      []usagestats.ModelUsageTrendPoint
	startTime   time.Time
	endTime     time.Time
	granularity string
	filters     usagestats.UsageLogFilters
	limit       int
}

func (s *modelTrendUsageRepoStub) GetModelUsageTrendWithUsageFilters(
	_ context.Context,
	startTime, endTime time.Time,
	granularity string,
	filters usagestats.UsageLogFilters,
	limit int,
) ([]usagestats.ModelUsageTrendPoint, error) {
	s.startTime = startTime
	s.endTime = endTime
	s.granularity = granularity
	s.filters = filters
	s.limit = limit
	return s.points, nil
}

func newModelTrendUsageRouter(repo *modelTrendUsageRepoStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	usageService := service.NewUsageService(repo, nil, nil, nil)
	handler := NewUsageHandler(usageService, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 42})
		c.Next()
	})
	router.GET("/usage/dashboard/model-trend", handler.DashboardModelTrend)
	return router
}

func TestDashboardModelTrendReturnsRequestedModelSeries(t *testing.T) {
	repo := &modelTrendUsageRepoStub{
		points: []usagestats.ModelUsageTrendPoint{
			{Date: "2026-07-06", Model: "gpt-5.6-sol", Requests: 10, ActualCost: 1.5},
			{Date: "2026-07-06", IsOther: true, Requests: 2, ActualCost: 0.2},
		},
	}
	router := newModelTrendUsageRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/model-trend?start_date=2026-07-06&end_date=2026-07-12&granularity=hour&timezone=Asia%2FShanghai", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data struct {
			Trend       []usagestats.ModelUsageTrendPoint `json:"trend"`
			Models      []string                          `json:"models"`
			StartDate   string                            `json:"start_date"`
			EndDate     string                            `json:"end_date"`
			Granularity string                            `json:"granularity"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Trend, 2)
	require.Equal(t, []string{"gpt-5.6-sol"}, body.Data.Models)
	require.Equal(t, "2026-07-06", body.Data.StartDate)
	require.Equal(t, "2026-07-12", body.Data.EndDate)
	require.Equal(t, "hour", body.Data.Granularity)
	require.Equal(t, int64(42), repo.filters.UserID)
	require.Equal(t, "hour", repo.granularity)
	require.Equal(t, 8, repo.limit)
	require.Equal(t, "Asia/Shanghai", req.URL.Query().Get("timezone"))
}

func TestDashboardModelTrendRejectsUnsupportedModelSource(t *testing.T) {
	repo := &modelTrendUsageRepoStub{}
	router := newModelTrendUsageRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/model-trend?model_source=upstream", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Empty(t, repo.granularity)
}

func TestDashboardModelTrendUsesExactRollingPeriodsAndRequestedGranularity(t *testing.T) {
	tests := []struct {
		period      string
		granularity string
		duration    time.Duration
	}{
		{period: "24h", granularity: "hour", duration: 24 * time.Hour},
		{period: "48h", granularity: "2h", duration: 48 * time.Hour},
		{period: "7d", granularity: "4h", duration: 7 * 24 * time.Hour},
		{period: "14d", granularity: "8h", duration: 14 * 24 * time.Hour},
		{period: "30d", granularity: "day", duration: 30 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			repo := &modelTrendUsageRepoStub{}
			router := newModelTrendUsageRouter(repo)
			before := time.Now()
			req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/model-trend?period="+tt.period+"&granularity="+tt.granularity+"&timezone=Asia%2FShanghai", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)
			after := time.Now()

			var body struct {
				Data struct {
					StartDate string `json:"start_date"`
					EndDate   string `json:"end_date"`
					StartTime string `json:"start_time"`
					EndTime   string `json:"end_time"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, tt.granularity, repo.granularity)
			require.Equal(t, tt.duration, repo.endTime.Sub(repo.startTime))
			require.False(t, repo.endTime.Before(before))
			require.False(t, repo.endTime.After(after))
			require.Equal(t, repo.startTime.Format("2006-01-02"), body.Data.StartDate)
			require.Equal(t, repo.endTime.Format("2006-01-02"), body.Data.EndDate)
			require.Equal(t, repo.startTime.Format(time.RFC3339), body.Data.StartTime)
			require.Equal(t, repo.endTime.Format(time.RFC3339), body.Data.EndTime)
		})
	}
}

func TestDashboardModelTrendUsesCalendarBoundariesForTodayAndYesterday(t *testing.T) {
	tests := []struct {
		period       string
		startDayDiff int
		endAtNow     bool
	}{
		{period: "today", startDayDiff: 0, endAtNow: true},
		{period: "yesterday", startDayDiff: -1, endAtNow: false},
	}

	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			repo := &modelTrendUsageRepoStub{}
			router := newModelTrendUsageRouter(repo)
			req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/model-trend?period="+tt.period+"&granularity=hour&timezone=Asia%2FShanghai", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			require.Zero(t, repo.startTime.Hour())
			require.Zero(t, repo.startTime.Minute())
			now := time.Now().In(repo.startTime.Location())
			expectedStart := time.Date(now.Year(), now.Month(), now.Day()+tt.startDayDiff, 0, 0, 0, 0, now.Location())
			require.Equal(t, expectedStart, repo.startTime)
			if tt.endAtNow {
				require.WithinDuration(t, now, repo.endTime, 2*time.Second)
			} else {
				require.Equal(t, expectedStart.AddDate(0, 0, 1), repo.endTime)
			}
		})
	}
}
