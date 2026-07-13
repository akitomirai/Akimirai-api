package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type userRequestLogRouteRepo struct {
	service.OpsRepository
}

func (r *userRequestLogRouteRepo) ListUserRequestLogs(
	_ context.Context,
	_ *service.UserRequestLogFilter,
) ([]*service.UserRequestLog, int64, error) {
	return []*service.UserRequestLog{}, 0, nil
}

func TestUserRoutesRegisterUnifiedRequestLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &userRequestLogRouteRepo{}
	opsService := service.NewOpsService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterUserRoutes(
		v1,
		&handler.Handlers{
			Usage: handler.NewUsageHandler(nil, nil, opsService, nil),
		},
		servermiddleware.JWTAuthMiddleware(func(c *gin.Context) {
			c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 42})
			c.Next()
		}),
		nil,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/usage/requests?kind=consumption", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
}
