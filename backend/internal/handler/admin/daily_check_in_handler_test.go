package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type dailyCheckInAdminRepoStub struct {
	filter service.DailyCheckInAdminFilter
	items  []service.DailyCheckInAdminRecord
	total  int64
}

func (r *dailyCheckInAdminRepoStub) GetForServiceDate(context.Context, int64, string) (*service.DailyCheckInRecord, float64, error) {
	return nil, 0, nil
}

func (r *dailyCheckInAdminRepoStub) Claim(context.Context, service.DailyCheckInClaim) (*service.DailyCheckInRecord, bool, error) {
	return nil, false, nil
}

func (r *dailyCheckInAdminRepoStub) ListForAdmin(_ context.Context, filter service.DailyCheckInAdminFilter) ([]service.DailyCheckInAdminRecord, int64, error) {
	r.filter = filter
	return r.items, r.total, nil
}

func (r *dailyCheckInAdminRepoStub) ListForUser(context.Context, int64, int, int) ([]service.DailyCheckInRecord, int64, error) {
	return nil, 0, nil
}

func TestDailyCheckInAdminHandlerListUsesPaginatedEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checkedAt := time.Date(2026, 7, 21, 3, 4, 5, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	repo := &dailyCheckInAdminRepoStub{
		items: []service.DailyCheckInAdminRecord{{
			ID: 41, UserID: 7, Email: "alice@example.com", Username: "Alice",
			ServiceDate: "2026-07-21", RewardAmount: 2, BalanceBefore: 10,
			BalanceAfter: 12, CheckedInAt: checkedAt, CreatedAt: checkedAt,
		}},
		total: 201,
	}
	h := NewDailyCheckInHandler(service.NewDailyCheckInService(repo, nil, nil))
	router := gin.New()
	router.GET("/api/v1/admin/daily-check-ins", h.List)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/daily-check-ins?page=2&page_size=500&q=%20alice%20&all=true", nil)
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	var body struct {
		Code int `json:"code"`
		Data struct {
			Items    []service.DailyCheckInAdminRecord `json:"items"`
			Total    int64                             `json:"total"`
			Page     int                               `json:"page"`
			PageSize int                               `json:"page_size"`
			Pages    int                               `json:"pages"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Zero(t, body.Code)
	require.Equal(t, int64(201), body.Data.Total)
	require.Equal(t, 2, body.Data.Page)
	require.Equal(t, 200, body.Data.PageSize)
	require.Equal(t, 2, body.Data.Pages)
	require.Len(t, body.Data.Items, 1)
	require.Equal(t, "alice@example.com", body.Data.Items[0].Email)
	require.Equal(t, "alice", repo.filter.Query)
	require.True(t, repo.filter.AllDates)
	require.Empty(t, repo.filter.ServiceDate)
}

func TestDailyCheckInAdminHandlerListRejectsInvalidServiceDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &dailyCheckInAdminRepoStub{}
	h := NewDailyCheckInHandler(service.NewDailyCheckInService(repo, nil, nil))
	router := gin.New()
	router.GET("/api/v1/admin/daily-check-ins", h.List)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/daily-check-ins?service_date=2026-02-30", nil)
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), "DAILY_CHECK_IN_SERVICE_DATE_INVALID")
}
