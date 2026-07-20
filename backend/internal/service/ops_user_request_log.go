package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type UserRequestLogKind string

const (
	UserRequestLogKindAll         UserRequestLogKind = "all"
	UserRequestLogKindConsumption UserRequestLogKind = "consumption"
	UserRequestLogKindError       UserRequestLogKind = "error"
)

type UserRequestLog struct {
	ID                  int64              `json:"id"`
	Kind                UserRequestLogKind `json:"kind"`
	CreatedAt           time.Time          `json:"created_at"`
	RequestID           string             `json:"request_id"`
	APIKeyID            *int64             `json:"api_key_id"`
	APIKeyName          string             `json:"api_key_name"`
	APIKeyDeleted       bool               `json:"api_key_deleted"`
	GroupID             *int64             `json:"group_id"`
	GroupName           string             `json:"group_name"`
	RateMultiplier      *float64           `json:"rate_multiplier"`
	Model               string             `json:"model"`
	ReasoningEffort     *string            `json:"reasoning_effort"`
	FirstTokenMs        *int               `json:"first_token_ms"`
	DurationMs          *int               `json:"duration_ms"`
	InputTokens         *int64             `json:"input_tokens"`
	OutputTokens        *int64             `json:"output_tokens"`
	CacheCreationTokens *int64             `json:"cache_creation_tokens"`
	CacheReadTokens     *int64             `json:"cache_read_tokens"`
	TotalTokens         *int64             `json:"total_tokens"`
	InputCost           *float64           `json:"input_cost"`
	OutputCost          *float64           `json:"output_cost"`
	CacheCreationCost   *float64           `json:"cache_creation_cost"`
	CacheReadCost       *float64           `json:"cache_read_cost"`
	TotalCost           *float64           `json:"total_cost"`
	ActualCost          *float64           `json:"actual_cost"`
	StatusCode          *int               `json:"status_code"`
	ErrorCode           *string            `json:"error_code"`
	ErrorMessage        *string            `json:"error_message"`

	RawErrorPhase   string `json:"-"`
	RawErrorType    string `json:"-"`
	RawErrorMessage string `json:"-"`
	RawErrorStream  bool   `json:"-"`
}

type UserRequestLogFilter struct {
	StartTime   time.Time
	EndTime     time.Time
	UserID      int64
	Kind        UserRequestLogKind
	APIKeyID    *int64
	GroupID     *int64
	Model       string
	RequestID   string
	SortBy      string
	SortOrder   string
	Page        int
	PageSize    int
	AllowErrors bool
}

func (f *UserRequestLogFilter) normalize() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	if f.Kind == "" {
		f.Kind = UserRequestLogKindAll
	}
	f.Model = strings.TrimSpace(f.Model)
	f.RequestID = strings.TrimSpace(f.RequestID)
	f.SortBy = strings.ToLower(strings.TrimSpace(f.SortBy))
	if f.SortBy != "duration_ms" {
		f.SortBy = "created_at"
	}
	f.SortOrder = strings.ToLower(strings.TrimSpace(f.SortOrder))
	if f.SortOrder != "asc" {
		f.SortOrder = "desc"
	}
}

func (s *OpsService) ListUserRequestLogs(
	ctx context.Context,
	userID int64,
	allowErrors bool,
	filter *UserRequestLogFilter,
) (*UserRequestLogList, error) {
	if userID <= 0 {
		return nil, infraerrors.Unauthorized("USER_NOT_AUTHENTICATED", "User not authenticated")
	}
	if filter == nil {
		filter = &UserRequestLogFilter{}
	}
	f := *filter
	f.UserID = userID
	f.AllowErrors = allowErrors
	f.normalize()

	switch f.Kind {
	case UserRequestLogKindAll, UserRequestLogKindConsumption, UserRequestLogKindError:
	default:
		return nil, infraerrors.BadRequest("INVALID_REQUEST_LOG_KIND", "Invalid request log kind")
	}
	if !allowErrors {
		if f.Kind == UserRequestLogKindError {
			return nil, infraerrors.Forbidden("USER_ERROR_VIEW_DISABLED", "Error requests view is disabled")
		}
		f.Kind = UserRequestLogKindConsumption
	}
	if s == nil || s.opsRepo == nil {
		return &UserRequestLogList{Items: []*UserRequestLog{}, Page: f.Page, PageSize: f.PageSize}, nil
	}

	items, total, err := s.opsRepo.ListUserRequestLogs(ctx, &f)
	if err != nil {
		return nil, fmt.Errorf("list user request logs: %w", err)
	}
	for _, item := range items {
		if item == nil || item.Kind != UserRequestLogKindError {
			continue
		}
		descriptor := UserErrorDescriptorForLog(&OpsErrorLog{
			Phase:      item.RawErrorPhase,
			Type:       item.RawErrorType,
			StatusCode: intValue(item.StatusCode),
			Message:    item.RawErrorMessage,
			Stream:     item.RawErrorStream,
		})
		code := string(descriptor.Code)
		message := descriptor.UserMessage
		item.ErrorCode = &code
		item.ErrorMessage = &message
	}

	return &UserRequestLogList{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

type UserRequestLogList struct {
	Items    []*UserRequestLog `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
