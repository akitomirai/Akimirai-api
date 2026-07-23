package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type dailyCheckInHandlerRepo struct {
	record  *service.DailyCheckInRecord
	created bool
}

func (r *dailyCheckInHandlerRepo) GetForServiceDate(context.Context, int64, string) (*service.DailyCheckInRecord, float64, error) {
	return r.record, r.record.BalanceAfter, nil
}

func (r *dailyCheckInHandlerRepo) Claim(context.Context, service.DailyCheckInClaim) (*service.DailyCheckInRecord, bool, error) {
	return r.record, r.created, nil
}

func (r *dailyCheckInHandlerRepo) ListForAdmin(context.Context, service.DailyCheckInAdminFilter) ([]service.DailyCheckInAdminRecord, int64, error) {
	return nil, 0, nil
}

func TestDailyCheckInHandlerClaimUsesAuthenticatedUserAndEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checkedAt := time.Date(2026, 7, 21, 3, 4, 5, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	repo := &dailyCheckInHandlerRepo{
		record: &service.DailyCheckInRecord{
			ID: 41, UserID: 7, ServiceDate: "2026-07-21", RewardAmount: 2,
			BalanceBefore: 10, BalanceAfter: 12, CheckedInAt: checkedAt,
		},
		created: true,
	}
	h := NewDailyCheckInHandler(service.NewDailyCheckInService(repo, nil, nil))
	router := gin.New()
	router.POST("/api/v1/user/check-in", func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 7})
		c.Next()
	}, h.Claim)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/user/check-in", nil))
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	var body struct {
		Code int `json:"code"`
		Data struct {
			CheckedIn        bool    `json:"checked_in"`
			AlreadyCheckedIn bool    `json:"already_checked_in"`
			RewardAmount     float64 `json:"reward_amount"`
			BalanceAfter     float64 `json:"balance_after"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Zero(t, body.Code)
	require.True(t, body.Data.CheckedIn)
	require.False(t, body.Data.AlreadyCheckedIn)
	require.Equal(t, float64(2), body.Data.RewardAmount)
	require.Equal(t, float64(12), body.Data.BalanceAfter)
}

func TestDailyCheckInHandlerRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDailyCheckInHandler(service.NewDailyCheckInService(&dailyCheckInHandlerRepo{}, nil, nil))
	router := gin.New()
	router.GET("/api/v1/user/check-in", h.GetStatus)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/user/check-in", nil))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
