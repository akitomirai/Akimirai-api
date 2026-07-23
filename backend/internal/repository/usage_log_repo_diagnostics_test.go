package repository

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPrepareUsageLogInsertIncludesDiagnosticsInCanonicalOrder(t *testing.T) {
	startedAt := time.Unix(1_750_001_000, 0).UTC()
	requestTotalMs := 900
	bodyReadMs := 15
	bodyBytes := int64(4096)
	connectionReused := false
	connectionReadyMs := 20
	dnsLookupMs := 2
	tcpConnectMs := 3
	tlsHandshakeMs := 4
	requestHeadersWrittenMs := 30
	requestWrittenMs := 40
	firstByteMs := 500
	responseHeadersReceivedMs := 505
	responseBodyFirstByteMs := 510
	firstEventMs := 520
	firstOutputCharacterMs := 530
	requestFirstTokenMs := 650
	routeKind := service.RequestRouteKindProxy
	proxyID := int64(22)
	proxyName := "jp-route"
	proxyProtocol := "socks5"
	fingerprint := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	status := 200
	promptCacheKeyHash := strings.Repeat("a", 64)
	promptCacheKeySource := string(service.PromptCacheKeySourceClientHeader)
	promptCachePrefixHash := strings.Repeat("b", 64)
	promptCacheToolsHash := strings.Repeat("c", 64)
	promptCacheSystemHash := strings.Repeat("d", 64)
	log := &service.UsageLog{
		UserID:                            1,
		APIKeyID:                          2,
		AccountID:                         3,
		RequestID:                         "req-diagnostics-order",
		Model:                             "gpt-5",
		RequestStartedAt:                  &startedAt,
		RequestTotalMs:                    &requestTotalMs,
		RequestBodyReadMs:                 &bodyReadMs,
		RequestBodyBytes:                  &bodyBytes,
		UpstreamConnectionReused:          &connectionReused,
		UpstreamConnectionReadyMs:         &connectionReadyMs,
		UpstreamDNSLookupMs:               &dnsLookupMs,
		UpstreamTCPConnectMs:              &tcpConnectMs,
		UpstreamTLSHandshakeMs:            &tlsHandshakeMs,
		UpstreamRequestHeadersWrittenMs:   &requestHeadersWrittenMs,
		UpstreamRequestWrittenMs:          &requestWrittenMs,
		UpstreamFirstByteMs:               &firstByteMs,
		UpstreamResponseHeadersReceivedMs: &responseHeadersReceivedMs,
		UpstreamResponseBodyFirstByteMs:   &responseBodyFirstByteMs,
		UpstreamFirstEventMs:              &firstEventMs,
		RequestFirstOutputCharacterMs:     &firstOutputCharacterMs,
		RequestFirstTokenMs:               &requestFirstTokenMs,
		RouteKind:                         &routeKind,
		ProxyIDSnapshot:                   &proxyID,
		ProxyNameSnapshot:                 &proxyName,
		ProxyProtocolSnapshot:             &proxyProtocol,
		RouteFingerprint:                  &fingerprint,
		FinalUpstreamStatus:               &status,
		RetryCount:                        2,
		AccountSwitchCount:                1,
		PromptCacheKeyHash:                &promptCacheKeyHash,
		PromptCacheKeySource:              &promptCacheKeySource,
		PromptCachePrefixHash:             &promptCachePrefixHash,
		PromptCacheToolsHash:              &promptCacheToolsHash,
		PromptCacheSystemHash:             &promptCacheSystemHash,
		AttemptTimeline: []service.RequestAttemptEvent{{
			Sequence:      1,
			AccountID:     3,
			Outcome:       "network_error",
			ErrorCategory: "network_error",
			Reason:        "failed via http://user:pass@proxy.example.test:8080/?key=secret",
		}},
		CreatedAt: time.Now().UTC(),
	}

	prepared := prepareUsageLogInsert(log)

	require.Len(t, prepared.args, len(usageLogInsertArgTypes))
	require.Equal(t, &startedAt, prepared.args[38])
	require.Equal(t, sql.NullInt64{Int64: int64(requestTotalMs), Valid: true}, prepared.args[39])
	require.Equal(t, sql.NullInt64{Int64: int64(bodyReadMs), Valid: true}, prepared.args[40])
	require.Equal(t, sql.NullInt64{Int64: bodyBytes, Valid: true}, prepared.args[41])
	require.Equal(t, sql.NullBool{Bool: connectionReused, Valid: true}, prepared.args[42])
	require.Equal(t, sql.NullInt64{Int64: int64(connectionReadyMs), Valid: true}, prepared.args[43])
	require.Equal(t, sql.NullInt64{Int64: int64(dnsLookupMs), Valid: true}, prepared.args[44])
	require.Equal(t, sql.NullInt64{Int64: int64(tcpConnectMs), Valid: true}, prepared.args[45])
	require.Equal(t, sql.NullInt64{Int64: int64(tlsHandshakeMs), Valid: true}, prepared.args[46])
	require.Equal(t, sql.NullInt64{Int64: int64(requestHeadersWrittenMs), Valid: true}, prepared.args[47])
	require.Equal(t, sql.NullInt64{Int64: int64(requestWrittenMs), Valid: true}, prepared.args[48])
	require.Equal(t, sql.NullInt64{Int64: int64(firstByteMs), Valid: true}, prepared.args[49])
	require.Equal(t, sql.NullInt64{Int64: int64(responseHeadersReceivedMs), Valid: true}, prepared.args[50])
	require.Equal(t, sql.NullInt64{Int64: int64(responseBodyFirstByteMs), Valid: true}, prepared.args[51])
	require.Equal(t, sql.NullInt64{Int64: int64(firstEventMs), Valid: true}, prepared.args[52])
	require.Equal(t, sql.NullInt64{Int64: int64(firstOutputCharacterMs), Valid: true}, prepared.args[53])
	require.Equal(t, sql.NullInt64{Int64: int64(requestFirstTokenMs), Valid: true}, prepared.args[54])
	require.Equal(t, sql.NullString{String: routeKind, Valid: true}, prepared.args[55])
	require.Equal(t, sql.NullInt64{Int64: proxyID, Valid: true}, prepared.args[56])
	require.Equal(t, sql.NullString{String: proxyName, Valid: true}, prepared.args[57])
	require.Equal(t, sql.NullString{String: proxyProtocol, Valid: true}, prepared.args[58])
	require.Equal(t, sql.NullString{String: fingerprint, Valid: true}, prepared.args[59])
	require.Equal(t, sql.NullInt64{Int64: int64(status), Valid: true}, prepared.args[60])
	require.Equal(t, 2, prepared.args[61])
	require.Equal(t, 1, prepared.args[62])
	timelineJSON, ok := prepared.args[63].(string)
	require.True(t, ok)
	require.NotContains(t, timelineJSON, "proxy.example.test")
	require.NotContains(t, timelineJSON, "user")
	require.NotContains(t, timelineJSON, "pass")
	require.NotContains(t, timelineJSON, "secret")
	require.Equal(t, sql.NullString{String: promptCacheKeyHash, Valid: true}, prepared.args[86])
	require.Equal(t, sql.NullString{String: promptCacheKeySource, Valid: true}, prepared.args[87])
	require.Equal(t, sql.NullString{String: promptCachePrefixHash, Valid: true}, prepared.args[88])
	require.Equal(t, sql.NullString{String: promptCacheToolsHash, Valid: true}, prepared.args[89])
	require.Equal(t, sql.NullString{String: promptCacheSystemHash, Valid: true}, prepared.args[90])
}

func TestRequestAttemptTimelineJSONRoundTripAndCap(t *testing.T) {
	events := make([]service.RequestAttemptEvent, service.RequestDiagnosticsAttemptLimit+4)
	for i := range events {
		events[i] = service.RequestAttemptEvent{Sequence: i + 1, AccountID: 7, Outcome: "http_error"}
	}

	raw, ok := nullRequestAttemptTimelineJSON(events).(string)
	require.True(t, ok)
	decoded := requestAttemptTimelineFromNullJSON(sql.NullString{String: raw, Valid: true})

	require.Len(t, decoded, service.RequestDiagnosticsAttemptLimit)
	require.Equal(t, 1, decoded[0].Sequence)
	require.Equal(t, service.RequestDiagnosticsAttemptLimit, decoded[len(decoded)-1].Sequence)
	require.Nil(t, requestAttemptTimelineFromNullJSON(sql.NullString{String: "{bad", Valid: true}))
}

func TestAppendUsageLogDiagnosticsWhereConditions(t *testing.T) {
	minTotal := 1000
	minFirstToken := 500
	minFirstByte := 250
	conditions, args := appendUsageLogDiagnosticsWhereConditions(nil, nil, UsageLogFilters{
		RouteKind:              service.RequestRouteKindProxy,
		ProxyID:                9,
		RetryOnly:              true,
		MinRequestTotalMs:      &minTotal,
		MinRequestFirstTokenMs: &minFirstToken,
		MinUpstreamFirstByteMs: &minFirstByte,
	})

	require.Equal(t, []string{
		"route_kind = $1",
		"proxy_id_snapshot = $2",
		"(retry_count > 0 OR account_switch_count > 0)",
		"request_total_ms >= $3",
		"request_first_token_ms >= $4",
		"upstream_first_byte_ms >= $5",
	}, conditions)
	require.Equal(t, []any{service.RequestRouteKindProxy, int64(9), 1000, 500, 250}, args)
	require.False(t, shouldUseFastUsageLogTotal(UsageLogFilters{RetryOnly: true}))
}

func TestUsageLogOrderByDiagnosticsUsesNullsLast(t *testing.T) {
	require.Equal(t, "request_total_ms DESC NULLS LAST, id DESC", usageLogOrderBy(pagination.PaginationParams{
		SortBy:    "request_total_ms",
		SortOrder: "desc",
	}))
	require.Equal(t, "request_started_at ASC NULLS LAST, id ASC", usageLogOrderBy(pagination.PaginationParams{
		SortBy:    "request_started_at",
		SortOrder: "asc",
	}))
}
