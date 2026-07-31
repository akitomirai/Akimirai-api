package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type dailyCheckInRouteRepo struct {
	record *service.DailyCheckInRecord
}

func (r *dailyCheckInRouteRepo) GetForServiceDate(context.Context, int64, string) (*service.DailyCheckInRecord, float64, error) {
	if r.record == nil {
		return nil, 10, nil
	}
	return r.record, r.record.BalanceAfter, nil
}

func (r *dailyCheckInRouteRepo) Claim(context.Context, service.DailyCheckInClaim) (*service.DailyCheckInRecord, bool, error) {
	return r.record, true, nil
}

func (r *dailyCheckInRouteRepo) ListForAdmin(context.Context, service.DailyCheckInAdminFilter) ([]service.DailyCheckInAdminRecord, int64, error) {
	return nil, 0, nil
}

func TestUserRoutesRegisterDailyCheckInGetAndPost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &dailyCheckInRouteRepo{record: &service.DailyCheckInRecord{
		ID: 1, UserID: 42, ServiceDate: "2026-07-21", RewardAmount: 2,
		BalanceBefore: 10, BalanceAfter: 12, CheckedInAt: time.Now(),
	}}
	checkInHandler := handler.NewDailyCheckInHandler(service.NewDailyCheckInService(repo, nil, nil))

	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterUserRoutes(
		v1,
		&handler.Handlers{DailyCheckIn: checkInHandler},
		servermiddleware.JWTAuthMiddleware(func(c *gin.Context) {
			c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
			c.Next()
		}),
		servermiddleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() }),
		nil,
		nil,
	)

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/api/v1/user/check-in", nil)
		router.ServeHTTP(recorder, request)
		require.Equal(t, http.StatusOK, recorder.Code, "%s: %s", method, recorder.Body.String())
	}
}
