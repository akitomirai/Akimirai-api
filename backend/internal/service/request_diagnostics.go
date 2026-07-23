package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	RequestDiagnosticsAttemptLimit = 32

	RequestRouteKindDirect = "direct"
	RequestRouteKindProxy  = "proxy"

	requestDiagnosticsReasonMaxLen = 256
)

var requestDiagnosticsURLRegex = regexp.MustCompile(`(?i)\b(?:https?|socks5h?|socks5)://[^\s]+`)

type requestDiagnosticsContextKey struct{}
type requestDiagnosticsAttemptContextKey struct{}

// RequestRouteSnapshot identifies an egress route without persisting its host or credentials.
type RequestRouteSnapshot struct {
	Kind          string `json:"kind"`
	ProxyID       *int64 `json:"proxy_id,omitempty"`
	ProxyName     string `json:"proxy_name,omitempty"`
	ProxyProtocol string `json:"proxy_protocol,omitempty"`
	Fingerprint   string `json:"fingerprint,omitempty"`
}

// RequestAttemptEvent is a bounded, sanitized timeline entry for one upstream HTTP attempt.
type RequestAttemptEvent struct {
	Sequence                  int                  `json:"sequence"`
	StartedMs                 int                  `json:"started_ms"`
	FinishedMs                *int                 `json:"finished_ms,omitempty"`
	ConnectionReused          *bool                `json:"connection_reused,omitempty"`
	ConnectionReadyMs         *int                 `json:"connection_ready_ms,omitempty"`
	DNSLookupMs               *int                 `json:"dns_lookup_ms,omitempty"`
	TCPConnectMs              *int                 `json:"tcp_connect_ms,omitempty"`
	TLSHandshakeMs            *int                 `json:"tls_handshake_ms,omitempty"`
	RequestHeadersWrittenMs   *int                 `json:"request_headers_written_ms,omitempty"`
	RequestWrittenMs          *int                 `json:"request_written_ms,omitempty"`
	FirstByteMs               *int                 `json:"first_byte_ms,omitempty"`
	ResponseHeadersReceivedMs *int                 `json:"response_headers_received_ms,omitempty"`
	ResponseBodyFirstByteMs   *int                 `json:"response_body_first_byte_ms,omitempty"`
	FirstEventMs              *int                 `json:"first_event_ms,omitempty"`
	FirstOutputCharacterMs    *int                 `json:"first_output_character_ms,omitempty"`
	AccountID                 int64                `json:"account_id"`
	AccountName               string               `json:"account_name,omitempty"`
	Route                     RequestRouteSnapshot `json:"route"`
	UpstreamStatus            *int                 `json:"upstream_status,omitempty"`
	Outcome                   string               `json:"outcome"`
	ErrorCategory             string               `json:"error_category,omitempty"`
	Reason                    string               `json:"reason,omitempty"`
}

// RequestDiagnosticsSnapshot is copied into the asynchronous usage-record task.
type RequestDiagnosticsSnapshot struct {
	RequestStartedAt                  time.Time
	RequestTotalMs                    *int
	RequestBodyReadMs                 *int
	RequestBodyBytes                  *int64
	UpstreamConnectionReused          *bool
	UpstreamConnectionReadyMs         *int
	UpstreamDNSLookupMs               *int
	UpstreamTCPConnectMs              *int
	UpstreamTLSHandshakeMs            *int
	UpstreamRequestHeadersWrittenMs   *int
	UpstreamRequestWrittenMs          *int
	UpstreamFirstByteMs               *int
	UpstreamResponseHeadersReceivedMs *int
	UpstreamResponseBodyFirstByteMs   *int
	UpstreamFirstEventMs              *int
	RequestFirstOutputCharacterMs     *int
	RequestFirstTokenMs               *int
	Route                             RequestRouteSnapshot
	FinalUpstreamStatus               *int
	RetryCount                        int
	AccountSwitchCount                int
	AttemptTimeline                   []RequestAttemptEvent
	PromptCache                       PromptCacheDiagnostics
}

type requestDiagnosticsSelectedAccount struct {
	name  string
	route RequestRouteSnapshot
}

// RequestDiagnostics is a request-scoped, fail-open collector for latency attribution.
type RequestDiagnostics struct {
	mu sync.Mutex

	startedAt time.Time

	requestBodyReadMs   *int
	requestBodyBytes    *int64
	requestFirstTokenMs *int
	promptCache         PromptCacheDiagnostics

	selectedAccounts map[int64]requestDiagnosticsSelectedAccount
	lastAccountID    int64
	hasAttempt       bool
	attemptCount     int
	retryCount       int
	accountSwitches  int
	attemptTimeline  []RequestAttemptEvent
	finalAttempt     *RequestAttemptEvent
}

// RequestDiagnosticsAttempt receives httptrace callbacks for one upstream request.
type RequestDiagnosticsAttempt struct {
	diagnostics *RequestDiagnostics
	once        sync.Once
	mu          sync.Mutex
	event       RequestAttemptEvent
	finished    bool

	dnsStartedAt     time.Time
	connectStartedAt time.Time
	tlsStartedAt     time.Time
}

func NewRequestDiagnostics(startedAt time.Time) *RequestDiagnostics {
	if startedAt.IsZero() {
		startedAt = time.Now()
	}
	return &RequestDiagnostics{
		startedAt:        startedAt,
		selectedAccounts: make(map[int64]requestDiagnosticsSelectedAccount),
		attemptTimeline:  make([]RequestAttemptEvent, 0, RequestDiagnosticsAttemptLimit),
	}
}

func WithRequestDiagnostics(ctx context.Context, diagnostics *RequestDiagnostics) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if diagnostics == nil {
		return ctx
	}
	return context.WithValue(ctx, requestDiagnosticsContextKey{}, diagnostics)
}

func RequestDiagnosticsFromContext(ctx context.Context) *RequestDiagnostics {
	if ctx == nil {
		return nil
	}
	diagnostics, _ := ctx.Value(requestDiagnosticsContextKey{}).(*RequestDiagnostics)
	return diagnostics
}

func WithRequestDiagnosticsAttempt(ctx context.Context, attempt *RequestDiagnosticsAttempt) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if attempt == nil {
		return ctx
	}
	return context.WithValue(ctx, requestDiagnosticsAttemptContextKey{}, attempt)
}

func RequestDiagnosticsAttemptFromContext(ctx context.Context) *RequestDiagnosticsAttempt {
	if ctx == nil {
		return nil
	}
	attempt, _ := ctx.Value(requestDiagnosticsAttemptContextKey{}).(*RequestDiagnosticsAttempt)
	return attempt
}

func (d *RequestDiagnostics) RecordBodyRead(duration time.Duration, bytes int64) {
	if d == nil || duration < 0 || bytes < 0 {
		return
	}
	durationMs := int(duration.Milliseconds())
	d.mu.Lock()
	d.requestBodyReadMs = &durationMs
	d.requestBodyBytes = &bytes
	d.mu.Unlock()
}

func (d *RequestDiagnostics) RecordPromptCacheDiagnostics(value PromptCacheDiagnostics) {
	if d == nil {
		return
	}
	value.Source = normalizePromptCacheKeySource(value.Source)
	d.mu.Lock()
	if value.KeyHash != "" || value.Source != PromptCacheKeySourceNone {
		d.promptCache.KeyHash = value.KeyHash
		d.promptCache.Source = value.Source
	} else if d.promptCache.Source == "" {
		d.promptCache.Source = PromptCacheKeySourceNone
	}
	if value.PrefixHash != "" {
		d.promptCache.PrefixHash = value.PrefixHash
	}
	if value.ToolsHash != "" {
		d.promptCache.ToolsHash = value.ToolsHash
	}
	if value.SystemHash != "" {
		d.promptCache.SystemHash = value.SystemHash
	}
	d.mu.Unlock()
}

func (d *RequestDiagnostics) SelectAccount(account *Account) {
	if d == nil || account == nil || account.ID <= 0 {
		return
	}
	d.mu.Lock()
	d.selectedAccounts[account.ID] = requestDiagnosticsSelectedAccount{
		name:  strings.TrimSpace(account.Name),
		route: RequestRouteSnapshotFromAccount(account),
	}
	d.mu.Unlock()
}

func (d *RequestDiagnostics) BeginHTTPAttemptAt(accountID int64, proxyURL string, at time.Time) *RequestDiagnosticsAttempt {
	if d == nil {
		return &RequestDiagnosticsAttempt{}
	}
	if at.IsZero() {
		at = time.Now()
	}

	d.mu.Lock()
	if d.hasAttempt {
		if d.lastAccountID == accountID {
			d.retryCount++
		} else {
			d.accountSwitches++
		}
	}
	d.hasAttempt = true
	d.lastAccountID = accountID
	d.attemptCount++
	sequence := d.attemptCount
	selected := d.selectedAccounts[accountID]
	d.mu.Unlock()

	route := requestRouteSnapshotFromRaw(proxyURL)
	if selected.route.Kind != "" && (proxyURL == "" || selected.route.Fingerprint == route.Fingerprint) {
		route = selected.route
	}

	return &RequestDiagnosticsAttempt{
		diagnostics: d,
		event: RequestAttemptEvent{
			Sequence:    sequence,
			StartedMs:   d.elapsedMs(at),
			AccountID:   accountID,
			AccountName: selected.name,
			Route:       cloneRequestRouteSnapshot(route),
		},
	}
}

func (a *RequestDiagnosticsAttempt) MarkDNSStartAt(at time.Time) {
	if a == nil || a.diagnostics == nil {
		return
	}
	a.mu.Lock()
	if a.dnsStartedAt.IsZero() || at.Before(a.dnsStartedAt) {
		a.dnsStartedAt = at
	}
	a.mu.Unlock()
}

// BeginConnectionAcquisitionAt starts the wire generation selected by net/http.
// A single RoundTrip can transparently retry and invoke GetConn more than once.
func (a *RequestDiagnosticsAttempt) BeginConnectionAcquisitionAt(_ time.Time) {
	if a == nil || a.diagnostics == nil {
		return
	}
	a.mu.Lock()
	a.dnsStartedAt = time.Time{}
	a.connectStartedAt = time.Time{}
	a.tlsStartedAt = time.Time{}
	a.event.ConnectionReused = nil
	a.event.ConnectionReadyMs = nil
	a.event.DNSLookupMs = nil
	a.event.TCPConnectMs = nil
	a.event.TLSHandshakeMs = nil
	a.event.RequestHeadersWrittenMs = nil
	a.event.RequestWrittenMs = nil
	a.event.FirstByteMs = nil
	a.event.ResponseHeadersReceivedMs = nil
	a.event.ResponseBodyFirstByteMs = nil
	a.event.FirstEventMs = nil
	a.event.FirstOutputCharacterMs = nil
	event, finished := a.finishedEventLocked()
	a.mu.Unlock()
	a.updateFinishedEvent(event, finished)
}

func (a *RequestDiagnosticsAttempt) MarkDNSDoneAt(at time.Time, err error) {
	if a == nil || a.diagnostics == nil || err != nil {
		return
	}
	a.mu.Lock()
	if a.event.DNSLookupMs == nil && !a.dnsStartedAt.IsZero() {
		value := nonNegativeDurationMs(a.dnsStartedAt, at)
		a.event.DNSLookupMs = &value
	}
	event, finished := a.finishedEventLocked()
	a.mu.Unlock()
	a.updateFinishedEvent(event, finished)
}

func (a *RequestDiagnosticsAttempt) MarkConnectStartAt(at time.Time) {
	if a == nil || a.diagnostics == nil {
		return
	}
	a.mu.Lock()
	if a.connectStartedAt.IsZero() || at.Before(a.connectStartedAt) {
		a.connectStartedAt = at
	}
	a.mu.Unlock()
}

func (a *RequestDiagnosticsAttempt) MarkConnectDoneAt(at time.Time, err error) {
	if a == nil || a.diagnostics == nil || err != nil {
		return
	}
	a.mu.Lock()
	if a.event.TCPConnectMs == nil && !a.connectStartedAt.IsZero() {
		value := nonNegativeDurationMs(a.connectStartedAt, at)
		a.event.TCPConnectMs = &value
	}
	event, finished := a.finishedEventLocked()
	a.mu.Unlock()
	a.updateFinishedEvent(event, finished)
}

func (a *RequestDiagnosticsAttempt) MarkTLSHandshakeStartAt(at time.Time) {
	if a == nil || a.diagnostics == nil {
		return
	}
	a.mu.Lock()
	if a.tlsStartedAt.IsZero() {
		a.tlsStartedAt = at
	}
	a.mu.Unlock()
}

func (a *RequestDiagnosticsAttempt) MarkTLSHandshakeDoneAt(at time.Time, err error) {
	if a == nil || a.diagnostics == nil || err != nil {
		return
	}
	a.mu.Lock()
	if a.event.TLSHandshakeMs == nil && !a.tlsStartedAt.IsZero() {
		value := nonNegativeDurationMs(a.tlsStartedAt, at)
		a.event.TLSHandshakeMs = &value
	}
	event, finished := a.finishedEventLocked()
	a.mu.Unlock()
	a.updateFinishedEvent(event, finished)
}

func (a *RequestDiagnosticsAttempt) MarkConnectionReadyAt(at time.Time, reused bool) {
	if a == nil || a.diagnostics == nil {
		return
	}
	value := a.diagnostics.elapsedMs(at)
	a.mu.Lock()
	if a.event.ConnectionReadyMs == nil {
		a.event.ConnectionReadyMs = &value
		a.event.ConnectionReused = cloneBool(&reused)
	}
	event, finished := a.finishedEventLocked()
	a.mu.Unlock()
	a.updateFinishedEvent(event, finished)
}

func (a *RequestDiagnosticsAttempt) MarkRequestHeadersWrittenAt(at time.Time) {
	a.markOffsetAt(at, func(event *RequestAttemptEvent, value *int) {
		if event.RequestHeadersWrittenMs == nil {
			event.RequestHeadersWrittenMs = value
		}
	})
}

func (a *RequestDiagnosticsAttempt) MarkRequestWrittenAt(at time.Time) {
	a.markOffsetAt(at, func(event *RequestAttemptEvent, value *int) {
		if event.RequestWrittenMs == nil {
			event.RequestWrittenMs = value
		}
	})
}

func (a *RequestDiagnosticsAttempt) MarkFirstResponseByteAt(at time.Time) {
	a.markOffsetAt(at, func(event *RequestAttemptEvent, value *int) {
		if event.FirstByteMs == nil {
			event.FirstByteMs = value
		}
	})
}

func (a *RequestDiagnosticsAttempt) MarkResponseHeadersReceivedAt(at time.Time) {
	a.markOffsetAt(at, func(event *RequestAttemptEvent, value *int) {
		if event.ResponseHeadersReceivedMs == nil {
			event.ResponseHeadersReceivedMs = value
		}
	})
}

func (a *RequestDiagnosticsAttempt) MarkResponseBodyFirstByteAt(at time.Time) {
	a.markOffsetAt(at, func(event *RequestAttemptEvent, value *int) {
		if event.ResponseBodyFirstByteMs == nil {
			event.ResponseBodyFirstByteMs = value
		}
	})
}

func (a *RequestDiagnosticsAttempt) MarkFirstStreamEventAt(at time.Time) {
	a.markOffsetAt(at, func(event *RequestAttemptEvent, value *int) {
		if event.FirstEventMs == nil {
			event.FirstEventMs = value
		}
	})
}

func (a *RequestDiagnosticsAttempt) MarkFirstOutputCharacterAt(at time.Time) {
	a.markOffsetAt(at, func(event *RequestAttemptEvent, value *int) {
		if event.FirstOutputCharacterMs == nil {
			event.FirstOutputCharacterMs = value
		}
	})
}

func (a *RequestDiagnosticsAttempt) FinishAt(at time.Time, upstreamStatus int, err error) {
	if a == nil || a.diagnostics == nil {
		return
	}
	a.once.Do(func() {
		finishedMs := a.diagnostics.elapsedMs(at)
		a.mu.Lock()
		a.event.FinishedMs = &finishedMs
		if upstreamStatus > 0 {
			a.event.UpstreamStatus = cloneInt(&upstreamStatus)
		}
		a.event.Outcome, a.event.ErrorCategory = requestDiagnosticsOutcome(upstreamStatus, err)
		if err != nil {
			a.event.Reason = sanitizeRequestDiagnosticReason(err.Error())
		}
		a.finished = true
		event := cloneRequestAttemptEvent(a.event)
		a.mu.Unlock()
		a.diagnostics.finishAttempt(event)
	})
}

func (a *RequestDiagnosticsAttempt) markOffsetAt(at time.Time, update func(*RequestAttemptEvent, *int)) {
	if a == nil || a.diagnostics == nil || update == nil {
		return
	}
	value := a.diagnostics.elapsedMs(at)
	a.mu.Lock()
	update(&a.event, &value)
	event, finished := a.finishedEventLocked()
	a.mu.Unlock()
	a.updateFinishedEvent(event, finished)
}

func (a *RequestDiagnosticsAttempt) finishedEventLocked() (RequestAttemptEvent, bool) {
	if !a.finished {
		return RequestAttemptEvent{}, false
	}
	return cloneRequestAttemptEvent(a.event), true
}

func (a *RequestDiagnosticsAttempt) updateFinishedEvent(event RequestAttemptEvent, finished bool) {
	if finished && a != nil && a.diagnostics != nil {
		a.diagnostics.updateFinishedAttempt(event)
	}
}

func (d *RequestDiagnostics) MarkFirstSemanticToken(forwardStartedAt time.Time, firstTokenMs *int) {
	if d == nil || firstTokenMs == nil || *firstTokenMs < 0 {
		return
	}
	value := d.elapsedMs(forwardStartedAt.Add(time.Duration(*firstTokenMs) * time.Millisecond))
	d.mu.Lock()
	d.requestFirstTokenMs = &value
	d.mu.Unlock()
}

func (d *RequestDiagnostics) SnapshotAt(completedAt time.Time, fallbackAccount *Account) RequestDiagnosticsSnapshot {
	if d == nil {
		return RequestDiagnosticsSnapshot{Route: RequestRouteSnapshotFromAccount(fallbackAccount)}
	}
	if completedAt.IsZero() {
		completedAt = time.Now()
	}
	totalMs := d.elapsedMs(completedAt)

	d.mu.Lock()
	snapshot := RequestDiagnosticsSnapshot{
		RequestStartedAt:    d.startedAt,
		RequestTotalMs:      cloneInt(&totalMs),
		RequestBodyReadMs:   cloneInt(d.requestBodyReadMs),
		RequestBodyBytes:    cloneInt64(d.requestBodyBytes),
		RequestFirstTokenMs: cloneInt(d.requestFirstTokenMs),
		RetryCount:          d.retryCount,
		AccountSwitchCount:  d.accountSwitches,
		PromptCache:         d.promptCache,
	}
	if d.finalAttempt != nil {
		snapshot.Route = cloneRequestRouteSnapshot(d.finalAttempt.Route)
		snapshot.UpstreamConnectionReused = cloneBool(d.finalAttempt.ConnectionReused)
		snapshot.UpstreamConnectionReadyMs = cloneInt(d.finalAttempt.ConnectionReadyMs)
		snapshot.UpstreamDNSLookupMs = cloneInt(d.finalAttempt.DNSLookupMs)
		snapshot.UpstreamTCPConnectMs = cloneInt(d.finalAttempt.TCPConnectMs)
		snapshot.UpstreamTLSHandshakeMs = cloneInt(d.finalAttempt.TLSHandshakeMs)
		snapshot.UpstreamRequestHeadersWrittenMs = cloneInt(d.finalAttempt.RequestHeadersWrittenMs)
		snapshot.UpstreamRequestWrittenMs = cloneInt(d.finalAttempt.RequestWrittenMs)
		snapshot.UpstreamFirstByteMs = cloneInt(d.finalAttempt.FirstByteMs)
		snapshot.UpstreamResponseHeadersReceivedMs = cloneInt(d.finalAttempt.ResponseHeadersReceivedMs)
		snapshot.UpstreamResponseBodyFirstByteMs = cloneInt(d.finalAttempt.ResponseBodyFirstByteMs)
		snapshot.UpstreamFirstEventMs = cloneInt(d.finalAttempt.FirstEventMs)
		snapshot.RequestFirstOutputCharacterMs = cloneInt(d.finalAttempt.FirstOutputCharacterMs)
		snapshot.FinalUpstreamStatus = cloneInt(d.finalAttempt.UpstreamStatus)
	}
	if snapshot.RetryCount > 0 || snapshot.AccountSwitchCount > 0 {
		snapshot.AttemptTimeline = make([]RequestAttemptEvent, len(d.attemptTimeline))
		for i := range d.attemptTimeline {
			snapshot.AttemptTimeline[i] = cloneRequestAttemptEvent(d.attemptTimeline[i])
		}
	}
	d.mu.Unlock()

	if snapshot.Route.Kind == "" {
		snapshot.Route = RequestRouteSnapshotFromAccount(fallbackAccount)
	}
	return snapshot
}

func (d *RequestDiagnostics) finishAttempt(event RequestAttemptEvent) {
	if d == nil {
		return
	}
	d.mu.Lock()
	if len(d.attemptTimeline) < RequestDiagnosticsAttemptLimit {
		d.attemptTimeline = append(d.attemptTimeline, cloneRequestAttemptEvent(event))
	}
	copy := cloneRequestAttemptEvent(event)
	d.finalAttempt = &copy
	d.mu.Unlock()
}

func (d *RequestDiagnostics) updateFinishedAttempt(event RequestAttemptEvent) {
	if d == nil || event.Sequence <= 0 {
		return
	}
	d.mu.Lock()
	for i := range d.attemptTimeline {
		if d.attemptTimeline[i].Sequence == event.Sequence {
			d.attemptTimeline[i] = cloneRequestAttemptEvent(event)
			break
		}
	}
	if d.finalAttempt != nil && d.finalAttempt.Sequence == event.Sequence {
		copy := cloneRequestAttemptEvent(event)
		d.finalAttempt = &copy
	}
	d.mu.Unlock()
}

func (d *RequestDiagnostics) elapsedMs(at time.Time) int {
	if d == nil {
		return 0
	}
	if at.IsZero() {
		at = time.Now()
	}
	value := at.Sub(d.startedAt).Milliseconds()
	if value < 0 {
		return 0
	}
	return int(value)
}

func RequestRouteSnapshotFromAccount(account *Account) RequestRouteSnapshot {
	if account == nil || account.ProxyID == nil || account.Proxy == nil {
		return RequestRouteSnapshot{Kind: RequestRouteKindDirect, Fingerprint: RequestRouteKindDirect}
	}
	snapshot := requestRouteSnapshotFromRaw(account.Proxy.URL())
	proxyID := *account.ProxyID
	snapshot.ProxyID = &proxyID
	snapshot.ProxyName = strings.TrimSpace(account.Proxy.Name)
	if protocol := strings.ToLower(strings.TrimSpace(account.Proxy.Protocol)); protocol != "" {
		snapshot.ProxyProtocol = protocol
	}
	return snapshot
}

func requestRouteSnapshotFromRaw(raw string) RequestRouteSnapshot {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RequestRouteSnapshot{Kind: RequestRouteKindDirect, Fingerprint: RequestRouteKindDirect}
	}

	canonical := raw
	protocol := ""
	if parsed, err := url.Parse(raw); err == nil {
		protocol = strings.ToLower(strings.TrimSpace(parsed.Scheme))
		host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
		port := strings.TrimSpace(parsed.Port())
		if host != "" {
			if port != "" {
				host = net.JoinHostPort(host, port)
			}
			canonical = protocol + "://" + host
		}
	}
	sum := sha256.Sum256([]byte(canonical))
	return RequestRouteSnapshot{
		Kind:          RequestRouteKindProxy,
		ProxyProtocol: protocol,
		Fingerprint:   hex.EncodeToString(sum[:]),
	}
}

func requestDiagnosticsOutcome(status int, err error) (string, string) {
	if err != nil {
		category := classifyRequestDiagnosticError(err)
		return category, category
	}
	if status >= 400 {
		return "http_error", "upstream_http_error"
	}
	if status > 0 {
		return "success", ""
	}
	return "unknown", ""
}

func classifyRequestDiagnosticError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	return "network_error"
}

func sanitizeRequestDiagnosticReason(reason string) string {
	reason = sanitizeUpstreamErrorMessage(strings.TrimSpace(reason))
	reason = requestDiagnosticsURLRegex.ReplaceAllString(reason, "[redacted-url]")
	return truncateString(strings.TrimSpace(reason), requestDiagnosticsReasonMaxLen)
}

// SanitizeRequestDiagnosticReason reapplies the persistence/API redaction boundary.
func SanitizeRequestDiagnosticReason(reason string) string {
	return sanitizeRequestDiagnosticReason(reason)
}

func cloneRequestAttemptEvent(event RequestAttemptEvent) RequestAttemptEvent {
	event.FinishedMs = cloneInt(event.FinishedMs)
	event.ConnectionReused = cloneBool(event.ConnectionReused)
	event.ConnectionReadyMs = cloneInt(event.ConnectionReadyMs)
	event.DNSLookupMs = cloneInt(event.DNSLookupMs)
	event.TCPConnectMs = cloneInt(event.TCPConnectMs)
	event.TLSHandshakeMs = cloneInt(event.TLSHandshakeMs)
	event.RequestHeadersWrittenMs = cloneInt(event.RequestHeadersWrittenMs)
	event.RequestWrittenMs = cloneInt(event.RequestWrittenMs)
	event.FirstByteMs = cloneInt(event.FirstByteMs)
	event.ResponseHeadersReceivedMs = cloneInt(event.ResponseHeadersReceivedMs)
	event.ResponseBodyFirstByteMs = cloneInt(event.ResponseBodyFirstByteMs)
	event.FirstEventMs = cloneInt(event.FirstEventMs)
	event.FirstOutputCharacterMs = cloneInt(event.FirstOutputCharacterMs)
	event.UpstreamStatus = cloneInt(event.UpstreamStatus)
	event.Route = cloneRequestRouteSnapshot(event.Route)
	return event
}

func cloneRequestRouteSnapshot(route RequestRouteSnapshot) RequestRouteSnapshot {
	route.ProxyID = cloneInt64(route.ProxyID)
	return route
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func nonNegativeDurationMs(start, end time.Time) int {
	value := end.Sub(start).Milliseconds()
	if value < 0 {
		return 0
	}
	return int(value)
}
