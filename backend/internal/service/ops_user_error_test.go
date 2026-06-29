package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestMapUserErrorCategory(t *testing.T) {
	cases := []struct {
		phase, etype, want string
	}{
		{"auth", "authentication_error", "auth"},
		{"request", "rate_limit_error", "rate_limit"},
		{"request", "billing_error", "quota"},
		{"request", "subscription_error", "quota"},
		{"request", "invalid_request_error", "invalid_request"},
		{"routing", "api_error", "service_unavailable"},
		{"upstream", "upstream_error", "upstream"},
		{"network", "api_error", "upstream"},
		{"internal", "api_error", "internal"},
		{"request", "cyber_policy_session_blocked", "cyber"},
		{"weird", "weird", "other"},
	}
	for _, c := range cases {
		if got := MapUserErrorCategory(c.phase, c.etype); got != c.want {
			t.Errorf("MapUserErrorCategory(%q,%q)=%q want %q", c.phase, c.etype, got, c.want)
		}
	}
}

func TestUserErrorDescriptorsAreComplete(t *testing.T) {
	required := []UserErrorCode{
		UserErrorCodeAuthInvalidKey,
		UserErrorCodeBalanceInsufficient,
		UserErrorCodeQuotaExceeded,
		UserErrorCodeModelDisabled,
		UserErrorCodeNoAvailableChannel,
		UserErrorCodeUpstreamAuthFailed,
		UserErrorCodeUpstreamRateLimited,
		UserErrorCodeUpstreamTimeout,
		UserErrorCodeUpstream5xx,
		UserErrorCodeStreamInterrupted,
		UserErrorCodeRequestInvalid,
		UserErrorCodePrivacyBlocked,
		UserErrorCodeContentBlocked,
	}

	for _, code := range required {
		desc := UserErrorDescriptorForCode(code)
		if desc.Code != code {
			t.Fatalf("descriptor code mismatch: got %q want %q", desc.Code, code)
		}
		if desc.UserMessage == "" || desc.AdminHint == "" || desc.Suggestion == "" || desc.HTTPStatus == 0 {
			t.Fatalf("descriptor %q incomplete: %+v", code, desc)
		}
	}
}

func TestCategoryToFilter(t *testing.T) {
	phases, types := CategoryToFilter("rate_limit")
	if len(types) != 1 || types[0] != "rate_limit_error" || len(phases) != 0 {
		t.Fatalf("rate_limit => phases=%v types=%v", phases, types)
	}
	phases, types = CategoryToFilter("auth")
	if len(phases) != 1 || phases[0] != "auth" || len(types) != 0 {
		t.Fatalf("auth => phases=%v types=%v", phases, types)
	}
	phases, types = CategoryToFilter("service_unavailable")
	if len(phases) != 1 || phases[0] != "routing" || len(types) != 0 {
		t.Fatalf("service_unavailable => phases=%v types=%v", phases, types)
	}
	phases, types = CategoryToFilter("upstream")
	if len(phases) != 2 || phases[0] != "upstream" || phases[1] != "network" || len(types) != 0 {
		t.Fatalf("upstream => phases=%v types=%v", phases, types)
	}
	phases, types = CategoryToFilter("internal")
	if len(phases) != 1 || phases[0] != "internal" || len(types) != 0 {
		t.Fatalf("internal => phases=%v types=%v", phases, types)
	}
	phases, types = CategoryToFilter("quota")
	if len(types) != 2 || types[0] != "billing_error" || types[1] != "subscription_error" || len(phases) != 0 {
		t.Fatalf("quota => phases=%v types=%v", phases, types)
	}
	phases, types = CategoryToFilter("invalid_request")
	if len(types) != 1 || types[0] != "invalid_request_error" || len(phases) != 0 {
		t.Fatalf("invalid_request => phases=%v types=%v", phases, types)
	}
	phases, types = CategoryToFilter("other")
	if len(phases) != 0 || len(types) != 0 {
		t.Fatalf("other => phases=%v types=%v", phases, types)
	}
}

func TestToUserErrorRequest_RedactsSensitiveFields(t *testing.T) {
	src := &OpsErrorLog{
		ID:              123,
		CreatedAt:       time.Unix(0, 0).UTC(),
		Model:           "m",
		RequestedModel:  "rm",
		InboundEndpoint: "/v1/chat/completions",
		StatusCode:      429,
		Platform:        "openai",
		Phase:           "request",
		Type:            "rate_limit_error",
		Message:         "rate limit exceeded",
		APIKeyName:      "my-key",
		APIKeyDeleted:   true,
	}
	out := ToUserErrorRequest(src)
	if out.ID != 123 {
		t.Errorf("want ID=123, got %d", out.ID)
	}
	if out.Model != "rm" {
		t.Errorf("want requested_model preferred, got %q", out.Model)
	}
	if out.Category != "rate_limit" {
		t.Errorf("category=%q", out.Category)
	}
	if out.StatusCode != 429 || out.InboundEndpoint != "/v1/chat/completions" || out.Platform != "openai" {
		t.Errorf("basic fields wrong: %+v", out)
	}
	if out.Message != "rate limit exceeded" {
		t.Errorf("want message=%q, got %q", "rate limit exceeded", out.Message)
	}
	if out.ErrorCode != string(UserErrorCodeQuotaExceeded) {
		t.Errorf("want error_code=%q, got %q", UserErrorCodeQuotaExceeded, out.ErrorCode)
	}
	if out.Explanation == "" {
		t.Error("expected non-empty explanation")
	}
	if out.Suggestion == "" {
		t.Error("expected non-empty suggestion")
	}
	if !out.Retryable {
		t.Error("expected retryable=true for rate limit")
	}
	if out.Charged {
		t.Error("expected charged=false for rate limit")
	}
	if out.HTTPStatus != 429 {
		t.Errorf("want http_status=429, got %d", out.HTTPStatus)
	}
	if out.KeyName != "my-key" {
		t.Errorf("want key_name=my-key, got %q", out.KeyName)
	}
	if !out.KeyDeleted {
		t.Error("want key_deleted=true")
	}
}

func TestToUserErrorRequestDetail_WhitelistAndRedacts(t *testing.T) {
	uid := int64(42)
	upstreamStatus := 503
	src := &OpsErrorLogDetail{
		OpsErrorLog: OpsErrorLog{
			ID:               999,
			CreatedAt:        time.Unix(1000, 0).UTC(),
			Model:            "gpt-4",
			RequestedModel:   "gpt-4-turbo",
			InboundEndpoint:  "/v1/chat/completions",
			StatusCode:       502,
			Platform:         "openai",
			Phase:            "upstream",
			Type:             "api_error",
			Message:          "upstream error",
			UserID:           &uid,
			UserEmail:        "secret@example.com",
			ClientIP:         func() *string { s := "1.2.3.4"; return &s }(),
			UpstreamEndpoint: "https://api.openai.com/v1/chat/completions",
		},
		ErrorBody:          `{"error":{"message":"upstream failed","type":"server_error"}}`,
		UserAgent:          "Mozilla/5.0 secret-agent",
		UpstreamStatusCode: &upstreamStatus,
	}

	out := ToUserErrorRequestDetail(src)
	if out == nil {
		t.Fatal("expected non-nil detail")
	}

	// 基础字段正确映射
	if out.ID != 999 {
		t.Errorf("want ID=999, got %d", out.ID)
	}
	if out.Message != "upstream error" {
		t.Errorf("want message=%q, got %q", "upstream error", out.Message)
	}
	if out.ErrorBody != src.ErrorBody {
		t.Errorf("ErrorBody mismatch")
	}
	if out.UpstreamStatusCode == nil || *out.UpstreamStatusCode != 503 {
		t.Errorf("UpstreamStatusCode mismatch")
	}

	// 序列化后不含敏感字段
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	raw := string(b)
	for _, forbidden := range []string{"user_email", "client_ip", "upstream_endpoint", "user_agent"} {
		if strings.Contains(raw, forbidden) {
			t.Errorf("sensitive field %q leaked in JSON output: %s", forbidden, raw)
		}
	}
	if strings.Contains(raw, "admin_hint") {
		t.Errorf("admin hint must not be exposed in user JSON: %s", raw)
	}
}

func TestClassifyUserErrorCode_CommonCodes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   *OpsErrorLog
		want UserErrorCode
	}{
		{name: "auth", in: &OpsErrorLog{Phase: "auth", Type: "authentication_error", StatusCode: 401}, want: UserErrorCodeAuthInvalidKey},
		{name: "balance", in: &OpsErrorLog{Phase: "request", Type: "billing_error", StatusCode: 403, Message: "insufficient balance"}, want: UserErrorCodeBalanceInsufficient},
		{name: "quota", in: &OpsErrorLog{Phase: "request", Type: "subscription_error", StatusCode: 429}, want: UserErrorCodeQuotaExceeded},
		{name: "model disabled", in: &OpsErrorLog{Phase: "request", Type: "invalid_request_error", StatusCode: 400, Message: "model not enabled"}, want: UserErrorCodeModelDisabled},
		{name: "no channel", in: &OpsErrorLog{Phase: "routing", Type: "api_error", StatusCode: 503}, want: UserErrorCodeNoAvailableChannel},
		{name: "upstream auth", in: &OpsErrorLog{Phase: "upstream", Type: "api_error", StatusCode: 401}, want: UserErrorCodeUpstreamAuthFailed},
		{name: "upstream rate", in: &OpsErrorLog{Phase: "upstream", Type: "api_error", StatusCode: 429}, want: UserErrorCodeUpstreamRateLimited},
		{name: "upstream timeout", in: &OpsErrorLog{Phase: "upstream", Type: "api_error", StatusCode: 504}, want: UserErrorCodeUpstreamTimeout},
		{name: "upstream 5xx", in: &OpsErrorLog{Phase: "upstream", Type: "api_error", StatusCode: 502}, want: UserErrorCodeUpstream5xx},
		{name: "stream", in: &OpsErrorLog{Phase: "upstream", Type: "api_error", StatusCode: 502, Stream: true, Message: "unexpected EOF"}, want: UserErrorCodeStreamInterrupted},
		{name: "invalid", in: &OpsErrorLog{Phase: "request", Type: "invalid_request_error", StatusCode: 400}, want: UserErrorCodeRequestInvalid},
		{name: "privacy", in: &OpsErrorLog{Phase: "request", Type: "cyber_policy", StatusCode: 403, Message: "privacy filter blocked pii"}, want: UserErrorCodePrivacyBlocked},
		{name: "content", in: &OpsErrorLog{Phase: "request", Type: "cyber_policy", StatusCode: 403, Message: "blocked"}, want: UserErrorCodeContentBlocked},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyUserErrorCode(tc.in); got != tc.want {
				t.Fatalf("ClassifyUserErrorCode()=%q want %q", got, tc.want)
			}
		})
	}
}

func TestExplainUserError_CommonOperationalCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   *OpsErrorLog
		want string
	}{
		{
			name: "upstream auth",
			in:   &OpsErrorLog{Phase: "upstream", Type: "upstream_error", StatusCode: 401},
			want: "credentials",
		},
		{
			name: "upstream rate limit",
			in:   &OpsErrorLog{Phase: "upstream", Type: "rate_limit_error", StatusCode: 429},
			want: "rate-limited",
		},
		{
			name: "stream interrupted",
			in:   &OpsErrorLog{Phase: "upstream", Type: "api_error", StatusCode: 502, Stream: true, Message: "unexpected EOF while reading stream"},
			want: "stream",
		},
		{
			name: "no healthy account",
			in:   &OpsErrorLog{Phase: "routing", Type: "api_error", StatusCode: 503},
			want: "No healthy upstream account",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExplainUserError(tc.in)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("ExplainUserError()=%q, want substring %q", got, tc.want)
			}
		})
	}
}

func TestToUserErrorRequestDetail_Nil(t *testing.T) {
	if out := ToUserErrorRequestDetail(nil); out != nil {
		t.Errorf("expected nil for nil input, got %+v", out)
	}
}
