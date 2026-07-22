package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// DailyCheckInHandler exposes the immutable check-in ledger to administrators.
type DailyCheckInHandler struct {
	service *service.DailyCheckInService
}

func NewDailyCheckInHandler(checkInService *service.DailyCheckInService) *DailyCheckInHandler {
	return &DailyCheckInHandler{service: checkInService}
}

// List handles GET /api/v1/admin/daily-check-ins.
func (h *DailyCheckInHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	allDates := false
	if raw := strings.TrimSpace(c.Query("all")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			response.BadRequest(c, "Invalid all flag")
			return
		}
		allDates = parsed
	}

	result, err := h.service.ListForAdmin(c.Request.Context(), service.DailyCheckInAdminFilter{
		Page:        page,
		PageSize:    pageSize,
		Query:       c.Query("q"),
		ServiceDate: c.Query("service_date"),
		AllDates:    allDates,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Items, result.Total, result.Page, result.PageSize)
}
