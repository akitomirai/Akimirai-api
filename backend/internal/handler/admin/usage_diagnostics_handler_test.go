package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type adminUsageDiagnosticsRepo struct {
	service.UsageLogRepository
	record *service.UsageLog
}

func (r *adminUsageDiagnosticsRepo) GetByID(context.Context, int64) (*service.UsageLog, error) {
	return r.record, nil
}

func TestAdminUsageDiagnosticsReturnsAdminOnlyTimeline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routeKind := service.RequestRouteKindProxy
	proxyName := "jp-egress"
	repo := &adminUsageDiagnosticsRepo{record: &service.UsageLog{
		ID:                42,
		RequestID:         "req-diagnostics",
		Model:             "gpt-5.6-sol",
		RouteKind:         &routeKind,
		ProxyNameSnapshot: &proxyName,
		RetryCount:        1,
		AttemptTimeline: []service.RequestAttemptEvent{{
			Sequence:  1,
			AccountID: 7,
			Outcome:   "network_error",
			Reason:    "failed via http://user:pass@proxy.example.test:8080/?key=secret",
			Route:     service.RequestRouteSnapshot{Kind: routeKind, ProxyName: proxyName},
		}},
	}}
	handler := NewUsageHandler(service.NewUsageService(repo, nil, nil, nil), nil, nil, nil)
	router := gin.New()
	router.GET("/admin/usage/:id/diagnostics", handler.Diagnostics)

	req := httptest.NewRequest(http.MethodGet, "/admin/usage/42/diagnostics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"request_id":"req-diagnostics"`)
	require.Contains(t, rec.Body.String(), `"proxy_name_snapshot":"jp-egress"`)
	require.Contains(t, rec.Body.String(), `"attempt_timeline"`)
	require.NotContains(t, rec.Body.String(), "proxy.example.test")
	require.NotContains(t, rec.Body.String(), "secret")
}

func TestAdminUsageDiagnosticsRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewUsageHandler(service.NewUsageService(&adminUsageDiagnosticsRepo{}, nil, nil, nil), nil, nil, nil)
	router := gin.New()
	router.GET("/admin/usage/:id/diagnostics", handler.Diagnostics)

	req := httptest.NewRequest(http.MethodGet, "/admin/usage/bad/diagnostics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
