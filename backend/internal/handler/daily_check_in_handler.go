package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DailyCheckInHandler struct {
	service *service.DailyCheckInService
}

func NewDailyCheckInHandler(checkInService *service.DailyCheckInService) *DailyCheckInHandler {
	return &DailyCheckInHandler{service: checkInService}
}

// GetStatus handles GET /api/v1/user/check-in.
func (h *DailyCheckInHandler) GetStatus(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	status, err := h.service.GetStatus(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
}

// Claim handles POST /api/v1/user/check-in.
func (h *DailyCheckInHandler) Claim(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	status, err := h.service.Claim(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
}
