package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// Diagnostics returns the administrator-only request diagnostic view for one usage row.
// GET /api/v1/admin/usage/:id/diagnostics
func (h *UsageHandler) Diagnostics(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid usage id")
		return
	}

	record, err := h.usageService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.UsageDiagnosticsFromServiceAdmin(record))
}

func applyAdminUsageDiagnosticFilters(c *gin.Context, filters *usagestats.UsageLogFilters) bool {
	if c == nil || filters == nil {
		return true
	}

	routeKind := strings.ToLower(strings.TrimSpace(c.Query("route_kind")))
	if routeKind != "" && routeKind != service.RequestRouteKindDirect && routeKind != service.RequestRouteKindProxy {
		response.BadRequest(c, "Invalid route_kind, allowed values: direct, proxy")
		return false
	}
	filters.RouteKind = routeKind

	if raw := strings.TrimSpace(c.Query("proxy_id")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			response.BadRequest(c, "Invalid proxy_id")
			return false
		}
		filters.ProxyID = value
	}

	if raw := strings.TrimSpace(c.Query("retry_only")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			response.BadRequest(c, "Invalid retry_only value, use true or false")
			return false
		}
		filters.RetryOnly = value
	}

	var ok bool
	if filters.MinRequestTotalMs, ok = parseNonNegativeUsageDuration(c, "min_request_total_ms"); !ok {
		return false
	}
	if filters.MinRequestFirstTokenMs, ok = parseNonNegativeUsageDuration(c, "min_request_first_token_ms"); !ok {
		return false
	}
	if filters.MinUpstreamFirstByteMs, ok = parseNonNegativeUsageDuration(c, "min_upstream_first_byte_ms"); !ok {
		return false
	}
	return true
}

func parseNonNegativeUsageDuration(c *gin.Context, key string) (*int, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		response.BadRequest(c, "Invalid "+key+", use a non-negative integer")
		return nil, false
	}
	return &value, true
}
