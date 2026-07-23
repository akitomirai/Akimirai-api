package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRequestRouteSnapshotFromAccountRedactsProxyCredentials(t *testing.T) {
	proxyID := int64(17)
	account := &Account{
		ID:      9,
		Name:    "primary",
		ProxyID: &proxyID,
		Proxy: &Proxy{
			ID:       proxyID,
			Name:     "jp-egress",
			Protocol: "http",
			Host:     "proxy.example.test",
			Port:     8080,
			Username: "diagnostic-user",
			Password: "diagnostic-password",
		},
	}

	snapshot := RequestRouteSnapshotFromAccount(account)

	require.Equal(t, RequestRouteKindProxy, snapshot.Kind)
	require.Equal(t, &proxyID, snapshot.ProxyID)
	require.Equal(t, "jp-egress", snapshot.ProxyName)
	require.Equal(t, "http", snapshot.ProxyProtocol)
	require.Len(t, snapshot.Fingerprint, 64)
	require.NotContains(t, snapshot.Fingerprint, "proxy.example.test")
	require.NotContains(t, snapshot.Fingerprint, "diagnostic-user")
	require.NotContains(t, snapshot.Fingerprint, "diagnostic-password")
}

func TestRequestDiagnosticsSingleAttemptKeepsTimelineNil(t *testing.T) {
	startedAt := time.Unix(1_750_000_000, 0)
	diagnostics := NewRequestDiagnostics(startedAt)
	diagnostics.RecordBodyRead(12*time.Millisecond, 321)
	diagnostics.SelectAccount(&Account{ID: 3, Name: "direct-account"})

	attempt := diagnostics.BeginHTTPAttemptAt(3, "", startedAt.Add(20*time.Millisecond))
	attempt.BeginConnectionAcquisitionAt(startedAt.Add(20 * time.Millisecond))
	attempt.MarkDNSStartAt(startedAt.Add(21 * time.Millisecond))
	attempt.MarkDNSDoneAt(startedAt.Add(23*time.Millisecond), nil)
	attempt.MarkConnectStartAt(startedAt.Add(23 * time.Millisecond))
	attempt.MarkConnectDoneAt(startedAt.Add(27*time.Millisecond), nil)
	attempt.MarkTLSHandshakeStartAt(startedAt.Add(27 * time.Millisecond))
	attempt.MarkTLSHandshakeDoneAt(startedAt.Add(32*time.Millisecond), nil)
	attempt.MarkConnectionReadyAt(startedAt.Add(32*time.Millisecond), false)
	attempt.MarkRequestHeadersWrittenAt(startedAt.Add(34 * time.Millisecond))
	attempt.MarkRequestWrittenAt(startedAt.Add(40 * time.Millisecond))
	attempt.MarkFirstResponseByteAt(startedAt.Add(70 * time.Millisecond))
	attempt.MarkResponseHeadersReceivedAt(startedAt.Add(72 * time.Millisecond))
	attempt.FinishAt(startedAt.Add(80*time.Millisecond), 200, nil)
	// Body and parser milestones arrive after http.Client.Do has returned.
	attempt.MarkResponseBodyFirstByteAt(startedAt.Add(75 * time.Millisecond))
	attempt.MarkFirstStreamEventAt(startedAt.Add(76 * time.Millisecond))
	attempt.MarkFirstOutputCharacterAt(startedAt.Add(78 * time.Millisecond))
	firstTokenMs := 35
	diagnostics.MarkFirstSemanticToken(startedAt.Add(20*time.Millisecond), &firstTokenMs)

	snapshot := diagnostics.SnapshotAt(startedAt.Add(100*time.Millisecond), nil)

	require.Equal(t, startedAt, snapshot.RequestStartedAt)
	require.Equal(t, diagIntPtr(100), snapshot.RequestTotalMs)
	require.Equal(t, diagIntPtr(12), snapshot.RequestBodyReadMs)
	require.Equal(t, diagInt64Ptr(321), snapshot.RequestBodyBytes)
	require.Equal(t, diagBoolPtr(false), snapshot.UpstreamConnectionReused)
	require.Equal(t, diagIntPtr(32), snapshot.UpstreamConnectionReadyMs)
	require.Equal(t, diagIntPtr(2), snapshot.UpstreamDNSLookupMs)
	require.Equal(t, diagIntPtr(4), snapshot.UpstreamTCPConnectMs)
	require.Equal(t, diagIntPtr(5), snapshot.UpstreamTLSHandshakeMs)
	require.Equal(t, diagIntPtr(34), snapshot.UpstreamRequestHeadersWrittenMs)
	require.Equal(t, diagIntPtr(40), snapshot.UpstreamRequestWrittenMs)
	require.Equal(t, diagIntPtr(70), snapshot.UpstreamFirstByteMs)
	require.Equal(t, diagIntPtr(72), snapshot.UpstreamResponseHeadersReceivedMs)
	require.Equal(t, diagIntPtr(75), snapshot.UpstreamResponseBodyFirstByteMs)
	require.Equal(t, diagIntPtr(76), snapshot.UpstreamFirstEventMs)
	require.Equal(t, diagIntPtr(78), snapshot.RequestFirstOutputCharacterMs)
	require.Equal(t, diagIntPtr(55), snapshot.RequestFirstTokenMs)
	require.Equal(t, diagIntPtr(200), snapshot.FinalUpstreamStatus)
	require.Equal(t, 0, snapshot.RetryCount)
	require.Equal(t, 0, snapshot.AccountSwitchCount)
	require.Nil(t, snapshot.AttemptTimeline)
	require.Equal(t, RequestRouteKindDirect, snapshot.Route.Kind)
}

func TestRequestDiagnosticsCountsRetryAndAccountSwitch(t *testing.T) {
	startedAt := time.Unix(1_750_000_100, 0)
	diagnostics := NewRequestDiagnostics(startedAt)
	proxyOneID := int64(11)
	proxyTwoID := int64(12)
	accountOne := &Account{ID: 1, Name: "account-one", ProxyID: &proxyOneID, Proxy: &Proxy{ID: proxyOneID, Name: "old-route", Protocol: "http", Host: "old.proxy.test", Port: 8080}}
	accountTwo := &Account{ID: 2, Name: "account-two", ProxyID: &proxyTwoID, Proxy: &Proxy{ID: proxyTwoID, Name: "jp-route", Protocol: "socks5", Host: "jp.proxy.test", Port: 1080}}

	diagnostics.SelectAccount(accountOne)
	first := diagnostics.BeginHTTPAttemptAt(accountOne.ID, accountOne.Proxy.URL(), startedAt.Add(10*time.Millisecond))
	first.MarkRequestWrittenAt(startedAt.Add(15 * time.Millisecond))
	first.FinishAt(startedAt.Add(30*time.Millisecond), 502, errors.New("Get http://user:pass@old.proxy.test:8080/path?key=secret: upstream reset"))

	diagnostics.SelectAccount(accountOne)
	second := diagnostics.BeginHTTPAttemptAt(accountOne.ID, accountOne.Proxy.URL(), startedAt.Add(40*time.Millisecond))
	second.FinishAt(startedAt.Add(55*time.Millisecond), 429, nil)

	diagnostics.SelectAccount(accountTwo)
	third := diagnostics.BeginHTTPAttemptAt(accountTwo.ID, accountTwo.Proxy.URL(), startedAt.Add(70*time.Millisecond))
	third.MarkRequestWrittenAt(startedAt.Add(75 * time.Millisecond))
	third.MarkFirstResponseByteAt(startedAt.Add(120 * time.Millisecond))
	third.FinishAt(startedAt.Add(150*time.Millisecond), 200, nil)

	snapshot := diagnostics.SnapshotAt(startedAt.Add(180*time.Millisecond), accountTwo)

	require.Equal(t, 1, snapshot.RetryCount)
	require.Equal(t, 1, snapshot.AccountSwitchCount)
	require.Len(t, snapshot.AttemptTimeline, 3)
	require.Equal(t, "network_error", snapshot.AttemptTimeline[0].ErrorCategory)
	require.NotContains(t, snapshot.AttemptTimeline[0].Reason, "user")
	require.NotContains(t, snapshot.AttemptTimeline[0].Reason, "pass")
	require.NotContains(t, snapshot.AttemptTimeline[0].Reason, "old.proxy.test")
	require.NotContains(t, snapshot.AttemptTimeline[0].Reason, "secret")
	require.Equal(t, "http_error", snapshot.AttemptTimeline[1].Outcome)
	require.Equal(t, "success", snapshot.AttemptTimeline[2].Outcome)
	require.Equal(t, accountTwo.ID, snapshot.AttemptTimeline[2].AccountID)
	require.Equal(t, &proxyTwoID, snapshot.Route.ProxyID)
	require.Equal(t, "jp-route", snapshot.Route.ProxyName)
	require.Equal(t, diagIntPtr(75), snapshot.UpstreamRequestWrittenMs)
	require.Equal(t, diagIntPtr(120), snapshot.UpstreamFirstByteMs)
	require.Equal(t, diagIntPtr(200), snapshot.FinalUpstreamStatus)
}

func TestRequestDiagnosticsCapsAttemptTimeline(t *testing.T) {
	startedAt := time.Unix(1_750_000_200, 0)
	diagnostics := NewRequestDiagnostics(startedAt)

	for i := 0; i < RequestDiagnosticsAttemptLimit+8; i++ {
		accountID := int64(i%2 + 1)
		diagnostics.SelectAccount(&Account{ID: accountID, Name: "account"})
		attempt := diagnostics.BeginHTTPAttemptAt(accountID, "", startedAt.Add(time.Duration(i)*time.Millisecond))
		attempt.FinishAt(startedAt.Add(time.Duration(i+1)*time.Millisecond), 503, nil)
	}

	snapshot := diagnostics.SnapshotAt(startedAt.Add(time.Second), nil)

	require.Len(t, snapshot.AttemptTimeline, RequestDiagnosticsAttemptLimit)
	require.Equal(t, RequestDiagnosticsAttemptLimit+7, snapshot.AccountSwitchCount)
}

func TestRequestDiagnosticsContextRoundTrip(t *testing.T) {
	diagnostics := NewRequestDiagnostics(time.Now())
	ctx := WithRequestDiagnostics(context.Background(), diagnostics)
	attempt := diagnostics.BeginHTTPAttemptAt(1, "", time.Now())
	ctx = WithRequestDiagnosticsAttempt(ctx, attempt)

	require.Same(t, diagnostics, RequestDiagnosticsFromContext(ctx))
	require.Same(t, attempt, RequestDiagnosticsAttemptFromContext(ctx))
	require.Nil(t, RequestDiagnosticsFromContext(nil))
	require.Nil(t, RequestDiagnosticsAttemptFromContext(nil))
}

func TestRequestDiagnosticsSnapshotIncludesPromptCacheHashes(t *testing.T) {
	diagnostics := NewRequestDiagnostics(time.Now())
	want := PromptCacheDiagnostics{
		KeyHash:    strings.Repeat("a", 64),
		Source:     PromptCacheKeySourceClientHeader,
		PrefixHash: strings.Repeat("b", 64),
		ToolsHash:  strings.Repeat("c", 64),
		SystemHash: strings.Repeat("d", 64),
	}

	diagnostics.RecordPromptCacheDiagnostics(want)
	got := diagnostics.SnapshotAt(time.Now(), nil)

	require.Equal(t, want, got.PromptCache)
}

func TestApplyRequestDiagnosticsCopiesPromptCacheHashesToUsageLog(t *testing.T) {
	diagnostics := &RequestDiagnosticsSnapshot{PromptCache: PromptCacheDiagnostics{
		KeyHash:    strings.Repeat("a", 64),
		Source:     PromptCacheKeySourceCompatDerived,
		PrefixHash: strings.Repeat("b", 64),
		ToolsHash:  strings.Repeat("c", 64),
		SystemHash: strings.Repeat("d", 64),
	}}
	usageLog := &UsageLog{}

	applyRequestDiagnosticsToUsageLog(usageLog, diagnostics, nil)

	require.Equal(t, diagnostics.PromptCache.KeyHash, *usageLog.PromptCacheKeyHash)
	require.Equal(t, string(PromptCacheKeySourceCompatDerived), *usageLog.PromptCacheKeySource)
	require.Equal(t, diagnostics.PromptCache.PrefixHash, *usageLog.PromptCachePrefixHash)
	require.Equal(t, diagnostics.PromptCache.ToolsHash, *usageLog.PromptCacheToolsHash)
	require.Equal(t, diagnostics.PromptCache.SystemHash, *usageLog.PromptCacheSystemHash)
}

func TestSanitizeRequestDiagnosticReasonBoundsAndRedacts(t *testing.T) {
	raw := "request failed via socks5://user:pass@proxy.example.test:1080/path?access_token=secret " + strings.Repeat("x", 400)
	safe := sanitizeRequestDiagnosticReason(raw)

	require.LessOrEqual(t, len(safe), requestDiagnosticsReasonMaxLen)
	require.NotContains(t, safe, "proxy.example.test")
	require.NotContains(t, safe, "user")
	require.NotContains(t, safe, "pass")
	require.NotContains(t, safe, "secret")
}

func diagIntPtr(value int) *int { return &value }

func diagInt64Ptr(value int64) *int64 { return &value }

func diagBoolPtr(value bool) *bool { return &value }
