package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type userRequestLogHandlerRepo struct {
	service.OpsRepository
	filter *service.UserRequestLogFilter
}

type userRequestLogSettingRepo struct {
	service.SettingRepository
	allowErrors bool
}

func (r *userRequestLogSettingRepo) GetMultiple(
	_ context.Context,
	_ []string,
) (map[string]string, error) {
	value := "false"
	if r.allowErrors {
		value = "true"
	}
	return map[string]string{service.SettingKeyAllowUserViewErrorRequests: value}, nil
}

func (s *userRequestLogHandlerRepo) ListUserRequestLogs(
	_ context.Context,
	filter *service.UserRequestLogFilter,
) ([]*service.UserRequestLog, int64, error) {
	copy := *filter
	s.filter = &copy
	return []*service.UserRequestLog{}, 0, nil
}

func newUserRequestLogHandlerRouter(
	repo *userRequestLogHandlerRepo,
	authenticated bool,
	allowErrors bool,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	opsService := service.NewOpsService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	settingService := service.NewSettingService(&userRequestLogSettingRepo{allowErrors: allowErrors}, nil)
	handler := NewUsageHandler(nil, nil, opsService, settingService)
	router := gin.New()
	if authenticated {
		router.Use(func(c *gin.Context) {
			c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 42})
			c.Next()
		})
	}
	router.GET("/usage/requests", handler.ListRequestLogs)
	return router
}

func TestUserRequestLogsAllFallsBackToConsumptionWhenErrorsDisabled(t *testing.T) {
	repo := &userRequestLogHandlerRepo{}
	router := newUserRequestLogHandlerRouter(repo, true, false)

	req := httptest.NewRequest(http.MethodGet, "/usage/requests?kind=all&start_date=2026-07-12&end_date=2026-07-13&page_size=500", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotNil(t, repo.filter)
	require.Equal(t, int64(42), repo.filter.UserID)
	require.Equal(t, service.UserRequestLogKindConsumption, repo.filter.Kind)
	require.Equal(t, 100, repo.filter.PageSize)
	require.False(t, repo.filter.AllowErrors)
}

func TestUserRequestLogsRejectsInvalidKindAndHiddenErrors(t *testing.T) {
	repo := &userRequestLogHandlerRepo{}
	router := newUserRequestLogHandlerRouter(repo, true, false)

	invalid := httptest.NewRecorder()
	router.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, "/usage/requests?kind=other", nil))
	require.Equal(t, http.StatusBadRequest, invalid.Code)

	hidden := httptest.NewRecorder()
	router.ServeHTTP(hidden, httptest.NewRequest(http.MethodGet, "/usage/requests?kind=error", nil))
	require.Equal(t, http.StatusForbidden, hidden.Code)
}

func TestUserRequestLogsAllowsErrorsWhenEnabled(t *testing.T) {
	repo := &userRequestLogHandlerRepo{}
	router := newUserRequestLogHandlerRouter(repo, true, true)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/usage/requests?kind=error", nil),
	)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.NotNil(t, repo.filter)
	require.Equal(t, service.UserRequestLogKindError, repo.filter.Kind)
	require.True(t, repo.filter.AllowErrors)
}

func TestUserRequestLogsRequiresAuthentication(t *testing.T) {
	router := newUserRequestLogHandlerRouter(&userRequestLogHandlerRepo{}, false, false)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/usage/requests", nil))
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
