package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type dailyCheckInAdminRouteRepo struct{}

func (dailyCheckInAdminRouteRepo) GetForServiceDate(context.Context, int64, string) (*service.DailyCheckInRecord, float64, error) {
	return nil, 0, nil
}

func (dailyCheckInAdminRouteRepo) Claim(context.Context, service.DailyCheckInClaim) (*service.DailyCheckInRecord, bool, error) {
	return nil, false, nil
}

func (dailyCheckInAdminRouteRepo) ListForAdmin(context.Context, service.DailyCheckInAdminFilter) ([]service.DailyCheckInAdminRecord, int64, error) {
	return []service.DailyCheckInAdminRecord{}, 0, nil
}

func (dailyCheckInAdminRouteRepo) ListForUser(context.Context, int64, int, int) ([]service.DailyCheckInRecord, int64, error) {
	return nil, 0, nil
}

func TestDailyCheckInAdminRouteRequiresAdminMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checkInHandler := adminhandler.NewDailyCheckInHandler(
		service.NewDailyCheckInService(dailyCheckInAdminRouteRepo{}, nil, nil),
	)
	handlers := &handler.Handlers{Admin: &handler.AdminHandlers{DailyCheckIn: checkInHandler}}

	router := gin.New()
	admin := router.Group("/api/v1/admin")
	admin.Use(func(c *gin.Context) {
		if c.GetHeader("X-Test-Role") != service.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"reason": "FORBIDDEN"})
			return
		}
		c.Next()
	})
	registerDailyCheckInAdminRoutes(admin, handlers)

	for _, tc := range []struct {
		name       string
		role       string
		wantStatus int
	}{
		{name: "unauthenticated", wantStatus: http.StatusForbidden},
		{name: "non_admin", role: service.RoleUser, wantStatus: http.StatusForbidden},
		{name: "admin", role: service.RoleAdmin, wantStatus: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/daily-check-ins?all=true", nil)
			if tc.role != "" {
				request.Header.Set("X-Test-Role", tc.role)
			}
			router.ServeHTTP(recorder, request)
			require.Equal(t, tc.wantStatus, recorder.Code, recorder.Body.String())
		})
	}
}
