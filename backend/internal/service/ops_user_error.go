package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// UserErrorRequest 是面向终端用户的"错误请求"精简脱敏视图（白名单）。
// 严禁包含 client_ip / user_agent / account / api_key_prefix / upstream_endpoint /
// user_email 等敏感或内部字段。注：message（网关标准化错误描述）与 key_name
// （用户自有 API Key 名称，KeysView 中本就可见）经产品决策对该用户开放；
// error_body 仅在详情接口（GetUserErrorRequestDetail）按归属校验后返回。
type UserErrorRequest struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Model           string    `json:"model"`
	InboundEndpoint string    `json:"inbound_endpoint"`
	StatusCode      int       `json:"status_code"`
	Category        string    `json:"category"`
	Platform        string    `json:"platform"`
	Message         string    `json:"message"`
	ErrorCode       string    `json:"error_code"`
	Explanation     string    `json:"explanation"`
	Suggestion      string    `json:"suggestion"`
	Retryable       bool      `json:"retryable"`
	Charged         bool      `json:"charged"`
	HTTPStatus      int       `json:"http_status"`
	KeyName         string    `json:"key_name"`
	KeyDeleted      bool      `json:"key_deleted"`
}

// UserErrorRequestList 是用户错误请求分页结果。
type UserErrorRequestList struct {
	Items    []*UserErrorRequest `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

type UserErrorCode string

const (
	UserErrorCodeAuthInvalidKey      UserErrorCode = "AUTH_INVALID_KEY"
	UserErrorCodeBalanceInsufficient UserErrorCode = "BALANCE_INSUFFICIENT"
	UserErrorCodeQuotaExceeded       UserErrorCode = "QUOTA_EXCEEDED"
	UserErrorCodeModelDisabled       UserErrorCode = "MODEL_DISABLED"
	UserErrorCodeNoAvailableChannel  UserErrorCode = "NO_AVAILABLE_CHANNEL"
	UserErrorCodeUpstreamAuthFailed  UserErrorCode = "UPSTREAM_AUTH_FAILED"
	UserErrorCodeUpstreamRateLimited UserErrorCode = "UPSTREAM_RATE_LIMITED"
	UserErrorCodeUpstreamTimeout     UserErrorCode = "UPSTREAM_TIMEOUT"
	UserErrorCodeUpstream5xx         UserErrorCode = "UPSTREAM_5XX"
	UserErrorCodeStreamInterrupted   UserErrorCode = "STREAM_INTERRUPTED"
	UserErrorCodeRequestInvalid      UserErrorCode = "REQUEST_FORMAT_INVALID"
	UserErrorCodePrivacyBlocked      UserErrorCode = "PRIVACY_BLOCKED"
	UserErrorCodeContentBlocked      UserErrorCode = "CONTENT_BLOCKED"
	UserErrorCodePlatformInternal    UserErrorCode = "PLATFORM_INTERNAL"
)

type UserErrorDescriptor struct {
	Code        UserErrorCode `json:"code"`
	UserMessage string        `json:"user_message"`
	AdminHint   string        `json:"admin_hint"`
	Retryable   bool          `json:"retryable"`
	Charged     bool          `json:"charged"`
	HTTPStatus  int           `json:"http_status"`
	Suggestion  string        `json:"suggestion"`
}

var userErrorDescriptors = map[UserErrorCode]UserErrorDescriptor{
	UserErrorCodeAuthInvalidKey: {
		Code:        UserErrorCodeAuthInvalidKey,
		UserMessage: "The request was rejected during authentication. Check that the API key is active and sent in the Authorization header.",
		AdminHint:   "Verify API key status, ownership, group binding, deleted-key audit attribution, and auth middleware logs.",
		HTTPStatus:  http.StatusUnauthorized,
		Suggestion:  "Create a new key or update the client Authorization header.",
	},
	UserErrorCodeBalanceInsufficient: {
		Code:        UserErrorCodeBalanceInsufficient,
		UserMessage: "The request was rejected because the account balance is insufficient.",
		AdminHint:   "Check balance ledger, recharge/payment order status, and any pending deduction reversals.",
		HTTPStatus:  http.StatusForbidden,
		Suggestion:  "Recharge the account or switch to an active subscription group.",
	},
	UserErrorCodeQuotaExceeded: {
		Code:        UserErrorCodeQuotaExceeded,
		UserMessage: "The request exceeded a quota, subscription limit, or platform rate limit.",
		AdminHint:   "Inspect API key quota, subscription windows, RPM limits, and user/group quota configuration.",
		Retryable:   true,
		HTTPStatus:  http.StatusTooManyRequests,
		Suggestion:  "Reduce concurrency, wait for the window to reset, or increase the configured quota.",
	},
	UserErrorCodeModelDisabled: {
		Code:        UserErrorCodeModelDisabled,
		UserMessage: "The requested model is not enabled for this key or route.",
		AdminHint:   "Check group model list, platform mapping, fallback route, and account capability settings.",
		HTTPStatus:  http.StatusBadRequest,
		Suggestion:  "Choose an enabled model from the model list or contact the operator to enable access.",
	},
	UserErrorCodeNoAvailableChannel: {
		Code:        UserErrorCodeNoAvailableChannel,
		UserMessage: "No healthy upstream account was available for this route.",
		AdminHint:   "Inspect channel health, account schedulability, proxy state, rate-limit freezes, and privacy requirements.",
		Retryable:   true,
		HTTPStatus:  http.StatusServiceUnavailable,
		Suggestion:  "Retry later or select another available model/group.",
	},
	UserErrorCodeUpstreamAuthFailed: {
		Code:        UserErrorCodeUpstreamAuthFailed,
		UserMessage: "The upstream provider rejected the selected account credentials or permissions.",
		AdminHint:   "Refresh or replace the upstream account credentials and verify provider-side scopes/organization access.",
		HTTPStatus:  http.StatusBadGateway,
		Suggestion:  "Retry later; if it persists, contact the operator to refresh the upstream account.",
	},
	UserErrorCodeUpstreamRateLimited: {
		Code:        UserErrorCodeUpstreamRateLimited,
		UserMessage: "The upstream provider rate-limited this account or model.",
		AdminHint:   "Check upstream 429/529 signals, account cooldown, model route concentration, and failover behavior.",
		Retryable:   true,
		HTTPStatus:  http.StatusTooManyRequests,
		Suggestion:  "Retry with backoff or switch to a less busy route.",
	},
	UserErrorCodeUpstreamTimeout: {
		Code:        UserErrorCodeUpstreamTimeout,
		UserMessage: "The upstream provider timed out before completing the request.",
		AdminHint:   "Check upstream latency, proxy quality, timeout settings, and request size.",
		Retryable:   true,
		HTTPStatus:  http.StatusGatewayTimeout,
		Suggestion:  "Retry later or reduce the request/response size.",
	},
	UserErrorCodeUpstream5xx: {
		Code:        UserErrorCodeUpstream5xx,
		UserMessage: "The upstream provider returned a temporary server error.",
		AdminHint:   "Inspect upstream status code/body, route failover, account health, and provider incident signals.",
		Retryable:   true,
		HTTPStatus:  http.StatusBadGateway,
		Suggestion:  "Retry later or use another model/account route.",
	},
	UserErrorCodeStreamInterrupted: {
		Code:        UserErrorCodeStreamInterrupted,
		UserMessage: "The stream ended before completion.",
		AdminHint:   "Check client disconnects, proxy resets, upstream stream errors, and partial usage records.",
		Retryable:   true,
		Charged:     true,
		HTTPStatus:  http.StatusBadGateway,
		Suggestion:  "Retry the request; if it repeats, reduce output size or use another route.",
	},
	UserErrorCodeRequestInvalid: {
		Code:        UserErrorCodeRequestInvalid,
		UserMessage: "The request format or parameters were invalid.",
		AdminHint:   "Check endpoint, model name, content schema, provider compatibility, and validation errors.",
		HTTPStatus:  http.StatusBadRequest,
		Suggestion:  "Review the request body, endpoint, model name, and SDK configuration.",
	},
	UserErrorCodePrivacyBlocked: {
		Code:        UserErrorCodePrivacyBlocked,
		UserMessage: "The request was blocked by the configured privacy policy.",
		AdminHint:   "Check privacy filter settings, group privacy requirement, and upstream account privacy mode.",
		HTTPStatus:  http.StatusForbidden,
		Suggestion:  "Remove blocked sensitive fields or choose a route that satisfies the privacy requirement.",
	},
	UserErrorCodeContentBlocked: {
		Code:        UserErrorCodeContentBlocked,
		UserMessage: "The request was blocked by the configured safety policy.",
		AdminHint:   "Review content moderation policy, cyber-policy decision, and user risk-control state.",
		HTTPStatus:  http.StatusForbidden,
		Suggestion:  "Modify the request content and retry within the platform policy.",
	},
	UserErrorCodePlatformInternal: {
		Code:        UserErrorCodePlatformInternal,
		UserMessage: "The gateway failed while processing this request.",
		AdminHint:   "Inspect gateway logs, dependency errors, panic recovery, and related request trace fields.",
		Retryable:   true,
		HTTPStatus:  http.StatusInternalServerError,
		Suggestion:  "Retry once, then contact support with the request time and model if it persists.",
	},
}

func UserErrorDescriptorForCode(code UserErrorCode) UserErrorDescriptor {
	if d, ok := userErrorDescriptors[code]; ok {
		return d
	}
	return userErrorDescriptors[UserErrorCodePlatformInternal]
}

func UserErrorDescriptorForLog(e *OpsErrorLog) UserErrorDescriptor {
	return UserErrorDescriptorForCode(ClassifyUserErrorCode(e))
}

// MapUserErrorCategory 把后端 error_phase + error_type 映射为用户侧粗分类码。
// 返回的是稳定的分类 code（前端做 i18n），不是展示文案。
func MapUserErrorCategory(phase, errType string) string {
	errType = strings.ToLower(strings.TrimSpace(errType))
	switch phase {
	case "auth":
		return "auth"
	case "routing":
		return "service_unavailable"
	case "upstream", "network":
		return "upstream"
	case "internal":
		return "internal"
	case "request":
		if strings.HasPrefix(errType, "cyber_policy") {
			return "cyber"
		}
		switch errType {
		case "rate_limit_error":
			return "rate_limit"
		case "billing_error", "subscription_error":
			return "quota"
		case "invalid_request_error":
			return "invalid_request"
		}
	}
	return "other"
}

// CategoryToFilter 把用户侧分类码反向映射为后端过滤条件（plain ANY）。
// 未知分类返回两个空切片（即不施加分类过滤）。
// 注意："other" 与未知分类都走 default 返回空切片——"other" 无对应的 phase/type 组合，无法精确反查，因此等价于不过滤。
func CategoryToFilter(category string) (phases []string, errorTypes []string) {
	switch category {
	case "auth":
		return []string{"auth"}, nil
	case "service_unavailable":
		return []string{"routing"}, nil
	case "upstream":
		return []string{"upstream", "network"}, nil
	case "internal":
		return []string{"internal"}, nil
	case "rate_limit":
		return nil, []string{"rate_limit_error"}
	case "quota":
		return nil, []string{"billing_error", "subscription_error"}
	case "invalid_request":
		return nil, []string{"invalid_request_error"}
	case "cyber":
		return []string{"request"}, []string{"cyber_policy", "cyber_policy_session_blocked"}
	default:
		return nil, nil
	}
}

// ToUserErrorRequest 把内部 OpsErrorLog 裁剪为用户安全视图。
func ToUserErrorRequest(e *OpsErrorLog) *UserErrorRequest {
	if e == nil {
		return nil
	}
	model := e.RequestedModel
	if model == "" {
		model = e.Model
	}
	descriptor := UserErrorDescriptorForLog(e)
	return &UserErrorRequest{
		ID:              e.ID,
		CreatedAt:       e.CreatedAt,
		Model:           model,
		InboundEndpoint: e.InboundEndpoint,
		StatusCode:      e.StatusCode,
		Category:        MapUserErrorCategory(e.Phase, e.Type),
		Platform:        e.Platform,
		Message:         e.Message,
		ErrorCode:       string(descriptor.Code),
		Explanation:     descriptor.UserMessage,
		Suggestion:      descriptor.Suggestion,
		Retryable:       descriptor.Retryable,
		Charged:         descriptor.Charged,
		HTTPStatus:      descriptor.HTTPStatus,
		KeyName:         e.APIKeyName,
		KeyDeleted:      e.APIKeyDeleted,
	}
}

func ClassifyUserErrorCode(e *OpsErrorLog) UserErrorCode {
	if e == nil {
		return UserErrorCodePlatformInternal
	}

	status := e.StatusCode
	msg := strings.ToLower(strings.TrimSpace(e.Message))
	errType := strings.ToLower(strings.TrimSpace(e.Type))
	category := MapUserErrorCategory(e.Phase, e.Type)

	if e.Stream && looksLikeStreamInterrupted(msg) {
		return UserErrorCodeStreamInterrupted
	}

	switch category {
	case "auth":
		return UserErrorCodeAuthInvalidKey
	case "quota":
		if errType == "billing_error" || looksLikeBalanceInsufficient(msg) {
			return UserErrorCodeBalanceInsufficient
		}
		return UserErrorCodeQuotaExceeded
	case "rate_limit":
		return UserErrorCodeQuotaExceeded
	case "invalid_request":
		if looksLikeModelDisabled(msg) {
			return UserErrorCodeModelDisabled
		}
		return UserErrorCodeRequestInvalid
	case "service_unavailable":
		return UserErrorCodeNoAvailableChannel
	case "upstream":
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			return UserErrorCodeUpstreamAuthFailed
		}
		if status == http.StatusTooManyRequests || status == 529 {
			return UserErrorCodeUpstreamRateLimited
		}
		if status == http.StatusRequestTimeout || status == http.StatusGatewayTimeout || looksLikeTimeout(msg) {
			return UserErrorCodeUpstreamTimeout
		}
		if looksLikeModelDisabled(msg) {
			return UserErrorCodeModelDisabled
		}
		if status >= 500 {
			return UserErrorCodeUpstream5xx
		}
		if status >= 400 {
			return UserErrorCodeRequestInvalid
		}
		return UserErrorCodeUpstream5xx
	case "internal":
		return UserErrorCodePlatformInternal
	case "cyber":
		if looksLikePrivacyBlocked(msg) {
			return UserErrorCodePrivacyBlocked
		}
		return UserErrorCodeContentBlocked
	default:
		return classifyGenericUserErrorCode(status, msg)
	}
}

func classifyGenericUserErrorCode(status int, msg string) UserErrorCode {
	if looksLikeModelDisabled(msg) {
		return UserErrorCodeModelDisabled
	}
	if looksLikePrivacyBlocked(msg) {
		return UserErrorCodePrivacyBlocked
	}
	if looksLikeStreamInterrupted(msg) {
		return UserErrorCodeStreamInterrupted
	}

	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return UserErrorCodeAuthInvalidKey
	case http.StatusTooManyRequests:
		return UserErrorCodeQuotaExceeded
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return UserErrorCodeUpstreamTimeout
	}
	if status >= 500 {
		return UserErrorCodePlatformInternal
	}
	if status >= 400 {
		return UserErrorCodeRequestInvalid
	}
	return UserErrorCodePlatformInternal
}

// ExplainUserError returns a stable, non-secret explanation that helps users
// understand what to try next without exposing upstream payloads or internals.
func ExplainUserError(e *OpsErrorLog) string {
	if e == nil {
		return ""
	}
	return UserErrorDescriptorForLog(e).UserMessage
}

func explainUpstreamUserError(status int, msg string, stream bool) string {
	if stream && looksLikeStreamInterrupted(msg) {
		return "The upstream stream ended before completion. Retry the request; if it repeats, switch account or model."
	}
	switch status {
	case 401, 403:
		return "The upstream provider rejected the selected account credentials or permissions. The operator may need to refresh or replace that upstream account."
	case 408, 504:
		return "The upstream provider timed out. Retry with a shorter request or try again later."
	case 429:
		return "The upstream provider rate-limited this account or model. Retry with backoff or wait for the upstream limit to reset."
	case 529:
		return "The upstream provider reported overload. Retry later or use another route."
	}
	if status >= 500 {
		return fmt.Sprintf("The upstream provider returned HTTP %d. Retry later or use another route.", status)
	}
	if status >= 400 {
		return fmt.Sprintf("The upstream provider rejected the request with HTTP %d. Check the model, endpoint, and upstream account permissions.", status)
	}
	if looksLikeStreamInterrupted(msg) {
		return "The upstream connection was interrupted before completion. Retry the request or use another route."
	}
	return "The upstream provider failed while handling this request. Retry later or use another route."
}

func explainGenericUserError(status int, msg string, stream bool) string {
	if stream && looksLikeStreamInterrupted(msg) {
		return "The stream was interrupted before completion. Retry the request or reduce the response size."
	}
	if status == 429 {
		return "The request was rate-limited. Retry with backoff."
	}
	if status >= 500 {
		return "The service encountered a temporary error. Retry later."
	}
	if status >= 400 {
		return "The request could not be processed. Check the request and retry."
	}
	return "The request failed. Retry later or contact support if it persists."
}

func looksLikeStreamInterrupted(msg string) bool {
	for _, needle := range []string{
		"stream",
		"eof",
		"unexpected eof",
		"context canceled",
		"client disconnected",
		"connection reset",
		"broken pipe",
		"timeout",
		"deadline exceeded",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func looksLikeBalanceInsufficient(msg string) bool {
	for _, needle := range []string{"balance", "insufficient", "not enough", "recharge", "余额"} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func looksLikeModelDisabled(msg string) bool {
	for _, needle := range []string{
		"model disabled",
		"model is disabled",
		"model not enabled",
		"model unavailable",
		"model is not available",
		"unsupported model",
		"unknown model",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func looksLikePrivacyBlocked(msg string) bool {
	for _, needle := range []string{"privacy", "training_off", "privacy_set", "redact", "pii"} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func looksLikeTimeout(msg string) bool {
	for _, needle := range []string{"timeout", "timed out", "deadline exceeded"} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

// UserErrorRequestDetail 是错误请求详情的脱敏视图(点击单行查看)。
// 在 UserErrorRequest 基础上额外暴露 error_body(上游错误响应正文)与 upstream_status_code;
// 仍严禁任何内部/敏感字段。
type UserErrorRequestDetail struct {
	UserErrorRequest
	ErrorBody          string `json:"error_body"`
	UpstreamStatusCode *int   `json:"upstream_status_code,omitempty"`
}

// ToUserErrorRequestDetail 把内部 OpsErrorLogDetail 裁剪为用户安全详情视图。
func ToUserErrorRequestDetail(e *OpsErrorLogDetail) *UserErrorRequestDetail {
	if e == nil {
		return nil
	}
	base := ToUserErrorRequest(&e.OpsErrorLog)
	return &UserErrorRequestDetail{
		UserErrorRequest:   *base,
		ErrorBody:          e.ErrorBody,
		UpstreamStatusCode: e.UpstreamStatusCode,
	}
}
