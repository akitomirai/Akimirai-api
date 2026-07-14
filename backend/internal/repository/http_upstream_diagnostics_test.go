package repository

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/servertiming"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestHTTPUpstreamDiagnosticsCapturesRequestWrittenAndFirstByte(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	diagnostics := service.NewRequestDiagnostics(time.Now())
	diagnostics.SelectAccount(&service.Account{ID: 41, Name: "direct"})
	req, err := http.NewRequestWithContext(
		service.WithRequestDiagnostics(t.Context(), diagnostics),
		http.MethodPost,
		server.URL,
		io.NopCloser(&zeroReader{}),
	)
	require.NoError(t, err)

	upstream := NewHTTPUpstream(&config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowPrivateHosts: true}}})
	resp, err := upstream.Do(req, "", 41, 1)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "ok", string(body))
	require.NoError(t, resp.Body.Close())

	snapshot := diagnostics.SnapshotAt(time.Now(), nil)
	require.Equal(t, diagBool(false), snapshot.UpstreamConnectionReused)
	require.NotNil(t, snapshot.UpstreamConnectionReadyMs)
	require.NotNil(t, snapshot.UpstreamTCPConnectMs)
	require.NotNil(t, snapshot.UpstreamRequestHeadersWrittenMs)
	require.NotNil(t, snapshot.UpstreamRequestWrittenMs)
	require.NotNil(t, snapshot.UpstreamFirstByteMs)
	require.NotNil(t, snapshot.UpstreamResponseHeadersReceivedMs)
	require.NotNil(t, snapshot.UpstreamResponseBodyFirstByteMs)
	require.GreaterOrEqual(t, *snapshot.UpstreamFirstByteMs, *snapshot.UpstreamRequestWrittenMs)
	require.GreaterOrEqual(t, *snapshot.UpstreamResponseHeadersReceivedMs, *snapshot.UpstreamFirstByteMs)
	require.GreaterOrEqual(t, *snapshot.UpstreamResponseBodyFirstByteMs, *snapshot.UpstreamResponseHeadersReceivedMs)
	require.Equal(t, http.StatusCreated, *snapshot.FinalUpstreamStatus)
	require.Nil(t, snapshot.AttemptTimeline)
}

func TestHTTPUpstreamDiagnosticsMarksPooledConnectionReuse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	upstream := NewHTTPUpstream(&config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowPrivateHosts: true}}})
	request := func(diagnostics *service.RequestDiagnostics) service.RequestDiagnosticsSnapshot {
		diagnostics.SelectAccount(&service.Account{ID: 52, Name: "direct"})
		req, err := http.NewRequestWithContext(service.WithRequestDiagnostics(t.Context(), diagnostics), http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := upstream.Do(req, "", 52, 1)
		require.NoError(t, err)
		_, err = io.Copy(io.Discard, resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		return diagnostics.SnapshotAt(time.Now(), nil)
	}

	first := request(service.NewRequestDiagnostics(time.Now()))
	second := request(service.NewRequestDiagnostics(time.Now()))

	require.Equal(t, diagBool(false), first.UpstreamConnectionReused)
	require.Equal(t, diagBool(true), second.UpstreamConnectionReused)
	require.Nil(t, second.UpstreamDNSLookupMs)
	require.Nil(t, second.UpstreamTCPConnectMs)
	require.Nil(t, second.UpstreamTLSHandshakeMs)
}

func TestHTTPUpstreamDiagnosticsAndServerTimingCoexist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	startedAt := time.Now()
	diagnostics := service.NewRequestDiagnostics(startedAt)
	diagnostics.SelectAccount(&service.Account{ID: 61, Name: "direct"})
	timing := servertiming.New(startedAt)
	ctx := service.WithRequestDiagnostics(t.Context(), diagnostics)
	ctx = servertiming.WithCollector(ctx, timing)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	upstream := NewHTTPUpstream(&config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowPrivateHosts: true}}})
	resp, err := upstream.Do(req, "", 61, 1)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	snapshot := diagnostics.SnapshotAt(time.Now(), nil)
	require.NotNil(t, snapshot.UpstreamRequestWrittenMs)
	require.NotNil(t, snapshot.UpstreamFirstByteMs)
	require.Contains(t, timing.HeaderValue(time.Now(), "bypass"), "dep_http;dur=")
}

func TestAttachHTTPUpstreamDiagnosticsWithoutCollectorIsNoOp(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)

	wrapped, attempt := attachHTTPUpstreamDiagnostics(req, "", 1)

	require.Same(t, req, wrapped)
	require.Nil(t, attempt)
}

type zeroReader struct{}

func (*zeroReader) Read([]byte) (int, error) { return 0, io.EOF }

func diagBool(value bool) *bool { return &value }
