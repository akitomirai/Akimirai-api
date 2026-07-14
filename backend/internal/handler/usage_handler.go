package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type userUsageFilters struct {
	Filters   usagestats.UsageLogFilters
	StartTime time.Time
	EndTime   time.Time
}

type userModelStat struct {
	Model               string  `json:"model"`
	Requests            int64   `json:"requests"`
	InputTokens         int64   `json:"input_tokens"`
	OutputTokens        int64   `json:"output_tokens"`
	CacheCreationTokens int64   `json:"cache_creation_tokens"`
	CacheReadTokens     int64   `json:"cache_read_tokens"`
	TotalTokens         int64   `json:"total_tokens"`
	Cost                float64 `json:"cost"`
	ActualCost          float64 `json:"actual_cost"`
}

type userGroupStat struct {
	GroupID     int64   `json:"group_id"`
	GroupName   string  `json:"group_name"`
	Requests    int64   `json:"requests"`
	TotalTokens int64   `json:"total_tokens"`
	Cost        float64 `json:"cost"`
	ActualCost  float64 `json:"actual_cost"`
}

type userAPIKeyStat struct {
	APIKeyID    int64   `json:"api_key_id"`
	APIKeyName  string  `json:"api_key_name"`
	Requests    int64   `json:"requests"`
	TotalTokens int64   `json:"total_tokens"`
	Cost        float64 `json:"cost"`
	ActualCost  float64 `json:"actual_cost"`
}

// UsageHandler handles usage-related requests
type UsageHandler struct {
	usageService   *service.UsageService
	apiKeyService  *service.APIKeyService
	opsService     *service.OpsService
	settingService *service.SettingService
}

func parseUserUsageExplicitRange(c *gin.Context) (*time.Time, *time.Time, bool) {
	userTZ := c.Query("timezone")
	startTimeValue := strings.TrimSpace(c.Query("start_time"))
	endTimeValue := strings.TrimSpace(c.Query("end_time"))
	if startTimeValue != "" || endTimeValue != "" {
		if startTimeValue == "" || endTimeValue == "" {
			response.BadRequest(c, "start_time and end_time are required together")
			return nil, nil, false
		}
		startTime, err := timezone.ParseInUserLocation("2006-01-02T15:04", startTimeValue, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_time format, use YYYY-MM-DDTHH:mm")
			return nil, nil, false
		}
		endTime, err := timezone.ParseInUserLocation("2006-01-02T15:04", endTimeValue, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_time format, use YYYY-MM-DDTHH:mm")
			return nil, nil, false
		}
		if !endTime.After(startTime) {
			response.BadRequest(c, "end_time must be after start_time")
			return nil, nil, false
		}
		return &startTime, &endTime, true
	}

	var startPtr, endPtr *time.Time
	if startDateValue := strings.TrimSpace(c.Query("start_date")); startDateValue != "" {
		startTime, err := timezone.ParseInUserLocation("2006-01-02", startDateValue, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
			return nil, nil, false
		}
		startPtr = &startTime
	}
	if endDateValue := strings.TrimSpace(c.Query("end_date")); endDateValue != "" {
		endTime, err := timezone.ParseInUserLocation("2006-01-02", endDateValue, userTZ)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, use YYYY-MM-DD")
			return nil, nil, false
		}
		endTime = endTime.AddDate(0, 0, 1)
		endPtr = &endTime
	}
	return startPtr, endPtr, true
}

// NewUsageHandler creates a new UsageHandler
func NewUsageHandler(
	usageService *service.UsageService,
	apiKeyService *service.APIKeyService,
	opsService *service.OpsService,
	settingService *service.SettingService,
) *UsageHandler {
	return &UsageHandler{
		usageService:   usageService,
		apiKeyService:  apiKeyService,
		opsService:     opsService,
		settingService: settingService,
	}
}

func (h *UsageHandler) parseUserUsageFilters(c *gin.Context, requireRange bool) (*userUsageFilters, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return nil, false
	}

	var apiKeyID int64
	if apiKeyIDStr := strings.TrimSpace(c.Query("api_key_id")); apiKeyIDStr != "" {
		id, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid api_key_id")
			return nil, false
		}
		if h.apiKeyService == nil {
			response.InternalError(c, "API key service not available")
			return nil, false
		}
		apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), id)
		if err != nil {
			response.ErrorFrom(c, err)
			return nil, false
		}
		if apiKey.UserID != subject.UserID {
			response.Forbidden(c, "Not authorized to access this API key's usage records")
			return nil, false
		}
		apiKeyID = id
	}

	var groupID int64
	if groupIDStr := strings.TrimSpace(c.Query("group_id")); groupIDStr != "" {
		id, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid group_id")
			return nil, false
		}
		groupID = id
	}

	var requestType *int16
	var stream *bool
	if requestTypeStr := strings.TrimSpace(c.Query("request_type")); requestTypeStr != "" {
		parsed, err := service.ParseUsageRequestType(requestTypeStr)
		if err != nil {
			response.BadRequest(c, err.Error())
			return nil, false
		}
		value := int16(parsed)
		requestType = &value
	} else if streamStr := strings.TrimSpace(c.Query("stream")); streamStr != "" {
		val, err := strconv.ParseBool(streamStr)
		if err != nil {
			response.BadRequest(c, "Invalid stream value, use true or false")
			return nil, false
		}
		stream = &val
	}

	var billingType *int8
	if billingTypeStr := strings.TrimSpace(c.Query("billing_type")); billingTypeStr != "" {
		val, err := strconv.ParseInt(billingTypeStr, 10, 8)
		if err != nil {
			response.BadRequest(c, "Invalid billing_type")
			return nil, false
		}
		bt := int8(val)
		billingType = &bt
	}

	billingMode := strings.TrimSpace(c.Query("billing_mode"))
	if billingMode != "" && !service.BillingMode(billingMode).IsValidUsageFilter() {
		response.BadRequest(c, "Invalid billing_mode")
		return nil, false
	}

	userTZ := c.Query("timezone")
	now := timezone.NowInUserLocation(userTZ)
	var startTime, endTime time.Time
	startPtr, endPtr, rangeOK := parseUserUsageExplicitRange(c)
	if !rangeOK {
		return nil, false
	}
	if startPtr != nil {
		startTime = *startPtr
	}
	if endPtr != nil {
		endTime = *endPtr
	}

	if requireRange {
		period := strings.TrimSpace(c.Query("period"))
		if startPtr == nil && endPtr == nil && period != "" {
			var valid bool
			startTime, endTime, valid = resolveUserUsagePeriodRange(now, userTZ, period)
			if !valid {
				response.BadRequest(c, "Invalid period")
				return nil, false
			}
			startPtr = &startTime
			endPtr = &endTime
		}
		if startPtr == nil {
			startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -7), userTZ)
			startPtr = &startTime
		}
		if endPtr == nil {
			endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
			endPtr = &endTime
		}
	}

	return &userUsageFilters{
		Filters: usagestats.UsageLogFilters{
			UserID:            subject.UserID,
			APIKeyID:          apiKeyID,
			GroupID:           groupID,
			Model:             strings.TrimSpace(c.Query("model")),
			ModelFilterSource: usagestats.ModelSourceRequested,
			RequestType:       requestType,
			Stream:            stream,
			BillingType:       billingType,
			BillingMode:       billingMode,
			StartTime:         startPtr,
			EndTime:           endPtr,
		},
		StartTime: derefTime(startPtr),
		EndTime:   derefTime(endPtr),
	}, true
}

func resolveUserUsagePeriodRange(now time.Time, userTZ, period string) (time.Time, time.Time, bool) {
	switch period {
	case "yesterday":
		end := timezone.StartOfDayInUserLocation(now, userTZ)
		return end.AddDate(0, 0, -1), end, true
	case "today":
		return timezone.StartOfDayInUserLocation(now, userTZ), now, true
	case "24h":
		return now.Add(-24 * time.Hour), now, true
	case "48h":
		return now.Add(-48 * time.Hour), now, true
	case "week", "7d":
		return now.Add(-7 * 24 * time.Hour), now, true
	case "14d":
		return now.Add(-14 * 24 * time.Hour), now, true
	case "month":
		return now.AddDate(0, -1, 0), now, true
	case "30d":
		return now.Add(-30 * 24 * time.Hour), now, true
	default:
		return time.Time{}, time.Time{}, false
	}
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func userUsageResponseEndDate(c *gin.Context, endTime time.Time) string {
	if strings.TrimSpace(c.Query("end_time")) != "" {
		return endTime.Format("2006-01-02")
	}
	switch strings.TrimSpace(c.Query("period")) {
	case "yesterday":
		return endTime.AddDate(0, 0, -1).Format("2006-01-02")
	case "today", "24h", "48h", "7d", "14d", "30d":
		return endTime.Format("2006-01-02")
	default:
		return endTime.AddDate(0, 0, -1).Format("2006-01-02")
	}
}

// List handles listing usage records with pagination
// GET /api/v1/usage
func (h *UsageHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	parsed, ok := h.parseUserUsageFilters(c, false)
	if !ok {
		return
	}

	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	records, result, err := h.usageService.ListWithFilters(c.Request.Context(), params, parsed.Filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UsageLog, 0, len(records))
	for i := range records {
		out = append(out, *dto.UsageLogFromService(&records[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

// ListRequestLogs returns one user-scoped, globally paginated projection across
// successful consumption rows and user-visible failed requests.
// GET /api/v1/usage/requests
func (h *UsageHandler) ListRequestLogs(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}

	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	if pageSize > 100 {
		pageSize = 100
	}

	kind := service.UserRequestLogKind(strings.ToLower(strings.TrimSpace(c.DefaultQuery("kind", "all"))))
	if kind != service.UserRequestLogKindAll &&
		kind != service.UserRequestLogKindConsumption &&
		kind != service.UserRequestLogKindError {
		response.BadRequest(c, "Invalid kind")
		return
	}
	allowErrors := h.settingService != nil && h.settingService.IsUserErrorViewAllowed(c.Request.Context())
	if kind == service.UserRequestLogKindError && !allowErrors {
		response.Forbidden(c, "Error requests view is disabled")
		return
	}

	var apiKeyID *int64
	if parsed.Filters.APIKeyID > 0 {
		value := parsed.Filters.APIKeyID
		apiKeyID = &value
	}
	var groupID *int64
	if parsed.Filters.GroupID > 0 {
		value := parsed.Filters.GroupID
		groupID = &value
	}

	filter := &service.UserRequestLogFilter{
		StartTime: parsed.StartTime,
		EndTime:   parsed.EndTime,
		Kind:      kind,
		APIKeyID:  apiKeyID,
		GroupID:   groupID,
		Model:     parsed.Filters.Model,
		RequestID: strings.TrimSpace(c.Query("request_id")),
		SortBy:    strings.TrimSpace(c.DefaultQuery("sort_by", "created_at")),
		SortOrder: strings.TrimSpace(c.DefaultQuery("sort_order", "desc")),
		Page:      page,
		PageSize:  pageSize,
	}
	if filter.SortBy != "created_at" && filter.SortBy != "duration_ms" {
		response.BadRequest(c, "Invalid sort_by")
		return
	}
	if filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		response.BadRequest(c, "Invalid sort_order")
		return
	}

	result, err := h.opsService.ListUserRequestLogs(
		c.Request.Context(),
		subject.UserID,
		allowErrors,
		filter,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Items, result.Total, result.Page, result.PageSize)
}

// ListErrors handles listing the current user's failed requests (redacted).
// GET /api/v1/usage/errors
func (h *UsageHandler) ListErrors(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Visibility switch (fail-closed). Defense-in-depth: frontend also hides the tab.
	if h.settingService == nil || !h.settingService.IsUserErrorViewAllowed(c.Request.Context()) {
		response.Forbidden(c, "Error requests view is disabled")
		return
	}
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 100 {
		pageSize = 100
	}

	filter := &service.OpsErrorLogFilter{Page: page, PageSize: pageSize}

	// Date range is a half-open interval. Custom minute precision and rolling
	// presets share the same semantics as the usage endpoints.
	startPtr, endPtr, rangeOK := parseUserUsageExplicitRange(c)
	if !rangeOK {
		return
	}
	if startPtr == nil && endPtr == nil {
		if period := strings.TrimSpace(c.Query("period")); period != "" {
			userTZ := c.Query("timezone")
			now := timezone.NowInUserLocation(userTZ)
			startTime, endTime, valid := resolveUserUsagePeriodRange(now, userTZ, period)
			if !valid {
				response.BadRequest(c, "Invalid period")
				return
			}
			startPtr = &startTime
			endPtr = &endTime
		}
	}
	filter.StartTime = startPtr
	filter.EndTime = endPtr

	filter.Model = strings.TrimSpace(c.Query("model"))

	if k := strings.TrimSpace(c.Query("api_key_id")); k != "" {
		n, err := strconv.ParseInt(k, 10, 64)
		if err != nil || n < 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		if n > 0 {
			filter.APIKeyID = &n
		}
	}

	if sc := strings.TrimSpace(c.Query("status_code")); sc != "" {
		n, err := strconv.Atoi(sc)
		if err != nil || n < 0 {
			response.BadRequest(c, "Invalid status_code")
			return
		}
		filter.StatusCodes = []int{n}
	}

	if cat := strings.TrimSpace(c.Query("category")); cat != "" {
		phases, types := service.CategoryToFilter(cat)
		filter.ErrorPhasesAny = phases
		filter.ErrorTypesAny = types
	}

	// 排序对齐用量明细:列白名单与方向归一在 repo 层,非法值回退 created_at DESC。
	filter.SetSort(c.Query("sort_by"), c.Query("sort_order"))

	result, err := h.opsService.ListUserErrorRequests(c.Request.Context(), subject.UserID, filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Items, int64(result.Total), result.Page, result.PageSize)
}

// GetErrorDetail handles fetching one of the current user's failed-request details (redacted).
// GET /api/v1/usage/errors/:id
func (h *UsageHandler) GetErrorDetail(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h.settingService == nil || !h.settingService.IsUserErrorViewAllowed(c.Request.Context()) {
		response.Forbidden(c, "Error requests view is disabled")
		return
	}
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
		return
	}
	detail, err := h.opsService.GetUserErrorRequestDetail(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, detail)
}

// GetByID handles getting a single usage record
// GET /api/v1/usage/:id
func (h *UsageHandler) GetByID(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	usageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid usage ID")
		return
	}

	record, err := h.usageService.GetByID(c.Request.Context(), usageID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	// 验证所有权
	if record.UserID != subject.UserID {
		response.Forbidden(c, "Not authorized to access this record")
		return
	}

	response.Success(c, dto.UsageLogFromService(record))
}

// Stats handles getting usage statistics
// GET /api/v1/usage/stats
func (h *UsageHandler) Stats(c *gin.Context) {
	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}

	stats, err := h.usageService.GetStatsWithFilters(c.Request.Context(), parsed.Filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	stats.TotalAccountCost = nil
	stats.UpstreamEndpoints = nil
	stats.EndpointPaths = nil

	response.Success(c, stats)
}

const (
	defaultAPIKeyDailyUsageDays = 30
	maxAPIKeyDailyUsageDays     = 90
)

func parseAPIKeyDailyUsageDays(raw string) (int, bool) {
	if strings.TrimSpace(raw) == "" {
		return defaultAPIKeyDailyUsageDays, true
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days <= 0 || days > maxAPIKeyDailyUsageDays {
		return 0, false
	}
	return days, true
}

func apiKeyDailyUsageRange(days int, userTZ string) (time.Time, time.Time) {
	now := timezone.NowInUserLocation(userTZ)
	startTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -(days-1)), userTZ)
	endTime := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
	return startTime, endTime
}

// DashboardStats handles getting user dashboard statistics
// GET /api/v1/usage/dashboard/stats
func (h *UsageHandler) DashboardStats(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	stats, err := h.usageService.GetUserDashboardStats(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, stats)
}

// DashboardTrend handles getting user usage trend data
// GET /api/v1/usage/dashboard/trend
func (h *UsageHandler) DashboardTrend(c *gin.Context) {
	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}
	granularity := normalizeUsageDashboardGranularity(c.DefaultQuery("granularity", "day"))

	trend, err := h.usageService.GetUsageTrendWithFilters(c.Request.Context(), parsed.StartTime, parsed.EndTime, granularity, parsed.Filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"trend":       trend,
		"start_date":  parsed.StartTime.Format("2006-01-02"),
		"end_date":    userUsageResponseEndDate(c, parsed.EndTime),
		"granularity": granularity,
	})
}

// DashboardModels handles getting user model usage statistics
// GET /api/v1/usage/dashboard/models
func (h *UsageHandler) DashboardModels(c *gin.Context) {
	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}

	modelSource := strings.TrimSpace(c.Query("model_source"))
	if modelSource != "" && modelSource != usagestats.ModelSourceRequested {
		response.BadRequest(c, "Invalid model_source, user usage only supports requested")
		return
	}

	stats, err := h.usageService.GetModelStatsWithFiltersBySource(c.Request.Context(), parsed.StartTime, parsed.EndTime, parsed.Filters, usagestats.ModelSourceRequested)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"models":     userModelStatsFromUsageStats(stats),
		"start_date": parsed.StartTime.Format("2006-01-02"),
		"end_date":   userUsageResponseEndDate(c, parsed.EndTime),
	})
}

// DashboardModelTrend handles requested-model usage trends for the user dashboard.
// GET /api/v1/usage/dashboard/model-trend
func (h *UsageHandler) DashboardModelTrend(c *gin.Context) {
	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}

	modelSource := strings.TrimSpace(c.Query("model_source"))
	if modelSource != "" && modelSource != usagestats.ModelSourceRequested {
		response.BadRequest(c, "Invalid model_source, user usage only supports requested")
		return
	}

	granularity := normalizeUsageDashboardGranularity(c.DefaultQuery("granularity", "day"))
	trend, err := h.usageService.GetModelUsageTrendWithFilters(c.Request.Context(), parsed.StartTime, parsed.EndTime, granularity, parsed.Filters, 8)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	models := make([]string, 0, 8)
	seen := make(map[string]struct{}, 8)
	for _, point := range trend {
		if point.IsOther || point.Model == "" {
			continue
		}
		if _, exists := seen[point.Model]; exists {
			continue
		}
		seen[point.Model] = struct{}{}
		models = append(models, point.Model)
	}
	endDate := userUsageResponseEndDate(c, parsed.EndTime)

	response.Success(c, gin.H{
		"trend":       trend,
		"models":      models,
		"start_date":  parsed.StartTime.Format("2006-01-02"),
		"end_date":    endDate,
		"start_time":  parsed.StartTime.Format(time.RFC3339),
		"end_time":    parsed.EndTime.Format(time.RFC3339),
		"granularity": granularity,
	})
}

// DashboardSnapshotV2 returns usage-page chart data scoped to the current user.
// GET /api/v1/usage/dashboard/snapshot-v2
func (h *UsageHandler) DashboardSnapshotV2(c *gin.Context) {
	parsed, ok := h.parseUserUsageFilters(c, true)
	if !ok {
		return
	}

	granularity := normalizeUsageDashboardGranularity(c.DefaultQuery("granularity", "day"))
	includeTrend, ok := parseBoolQueryWithDefault(c, "include_trend", true)
	if !ok {
		return
	}
	includeModels, ok := parseBoolQueryWithDefault(c, "include_model_stats", true)
	if !ok {
		return
	}
	includeGroups, ok := parseBoolQueryWithDefault(c, "include_group_stats", false)
	if !ok {
		return
	}
	includeAPIKeys, ok := parseBoolQueryWithDefault(c, "include_api_key_stats", false)
	if !ok {
		return
	}

	resp := gin.H{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"start_date":   parsed.StartTime.Format("2006-01-02"),
		"end_date":     userUsageResponseEndDate(c, parsed.EndTime),
		"granularity":  granularity,
	}

	if includeTrend {
		trend, err := h.usageService.GetUsageTrendWithFilters(c.Request.Context(), parsed.StartTime, parsed.EndTime, granularity, parsed.Filters)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		resp["trend"] = trend
	}
	if includeModels {
		models, err := h.usageService.GetModelStatsWithFiltersBySource(c.Request.Context(), parsed.StartTime, parsed.EndTime, parsed.Filters, usagestats.ModelSourceRequested)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		resp["models"] = userModelStatsFromUsageStats(models)
	}
	if includeGroups {
		groups, err := h.usageService.GetGroupStatsWithFilters(c.Request.Context(), parsed.StartTime, parsed.EndTime, parsed.Filters)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		resp["groups"] = userGroupStatsFromUsageStats(groups)
	}
	if includeAPIKeys {
		apiKeys, err := h.usageService.GetAPIKeyStatsWithFilters(c.Request.Context(), parsed.StartTime, parsed.EndTime, parsed.Filters)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		resp["api_keys"] = userAPIKeyStatsFromUsageStats(apiKeys)
	}

	response.Success(c, resp)
}

func normalizeUsageDashboardGranularity(raw string) string {
	switch strings.TrimSpace(raw) {
	case "hour", "2h", "4h", "8h", "day":
		return strings.TrimSpace(raw)
	default:
		return "day"
	}
}

func userAPIKeyStatsFromUsageStats(stats []usagestats.APIKeyStat) []userAPIKeyStat {
	out := make([]userAPIKeyStat, 0, len(stats))
	for _, stat := range stats {
		out = append(out, userAPIKeyStat{
			APIKeyID:    stat.APIKeyID,
			APIKeyName:  stat.APIKeyName,
			Requests:    stat.Requests,
			TotalTokens: stat.TotalTokens,
			Cost:        stat.Cost,
			ActualCost:  stat.ActualCost,
		})
	}
	return out
}

func userModelStatsFromUsageStats(stats []usagestats.ModelStat) []userModelStat {
	out := make([]userModelStat, 0, len(stats))
	for _, stat := range stats {
		out = append(out, userModelStat{
			Model:               stat.Model,
			Requests:            stat.Requests,
			InputTokens:         stat.InputTokens,
			OutputTokens:        stat.OutputTokens,
			CacheCreationTokens: stat.CacheCreationTokens,
			CacheReadTokens:     stat.CacheReadTokens,
			TotalTokens:         stat.TotalTokens,
			Cost:                stat.Cost,
			ActualCost:          stat.ActualCost,
		})
	}
	return out
}

func userGroupStatsFromUsageStats(stats []usagestats.GroupStat) []userGroupStat {
	out := make([]userGroupStat, 0, len(stats))
	for _, stat := range stats {
		out = append(out, userGroupStat{
			GroupID:     stat.GroupID,
			GroupName:   stat.GroupName,
			Requests:    stat.Requests,
			TotalTokens: stat.TotalTokens,
			Cost:        stat.Cost,
			ActualCost:  stat.ActualCost,
		})
	}
	return out
}

func parseBoolQueryWithDefault(c *gin.Context, key string, fallback bool) (bool, bool) {
	raw := c.Query(key)
	if strings.TrimSpace(raw) == "" {
		return fallback, true
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		response.BadRequest(c, "Invalid "+key+" value, use true or false")
		return false, false
	}
	return parsed, true
}

// BatchAPIKeysUsageRequest represents the request for batch API keys usage
type BatchAPIKeysUsageRequest struct {
	APIKeyIDs []int64 `json:"api_key_ids" binding:"required"`
}

// DashboardAPIKeysUsage handles getting usage stats for user's own API keys
// POST /api/v1/usage/dashboard/api-keys-usage
func (h *UsageHandler) DashboardAPIKeysUsage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req BatchAPIKeysUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if len(req.APIKeyIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	// Limit the number of API key IDs to prevent SQL parameter overflow
	if len(req.APIKeyIDs) > 100 {
		response.BadRequest(c, "Too many API key IDs (maximum 100 allowed)")
		return
	}

	validAPIKeyIDs, err := h.apiKeyService.VerifyOwnership(c.Request.Context(), subject.UserID, req.APIKeyIDs)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if len(validAPIKeyIDs) == 0 {
		response.Success(c, gin.H{"stats": map[string]any{}})
		return
	}

	stats, err := h.usageService.GetBatchAPIKeyUsageStats(c.Request.Context(), validAPIKeyIDs, time.Time{}, time.Time{})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"stats": stats})
}

// GetMyAPIKeyDailyUsage handles getting daily usage details for the current user's API key.
// GET /api/v1/user/api-keys/:id/usage/daily?days=30
func (h *UsageHandler) GetMyAPIKeyDailyUsage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	apiKeyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid API key ID")
		return
	}

	days, ok := parseAPIKeyDailyUsageDays(c.DefaultQuery("days", ""))
	if !ok {
		response.BadRequest(c, "Invalid days, allowed range is 1-90")
		return
	}

	if h.apiKeyService == nil {
		response.InternalError(c, "API key service is not configured")
		return
	}

	apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), apiKeyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if apiKey.UserID != subject.UserID {
		response.Forbidden(c, "Not authorized to access this API key's usage")
		return
	}

	userTZ := c.Query("timezone")
	startTime, endTime := apiKeyDailyUsageRange(days, userTZ)
	items, err := h.usageService.GetAPIKeyDailyUsage(c.Request.Context(), subject.UserID, apiKeyID, startTime, endTime)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"items":      items,
		"days":       days,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.AddDate(0, 0, -1).Format("2006-01-02"),
	})
}
