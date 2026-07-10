package repository

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
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
	require.NoError(t, resp.Body.Close())

	snapshot := diagnostics.SnapshotAt(time.Now(), nil)
	require.NotNil(t, snapshot.UpstreamRequestWrittenMs)
	require.NotNil(t, snapshot.UpstreamFirstByteMs)
	require.GreaterOrEqual(t, *snapshot.UpstreamFirstByteMs, *snapshot.UpstreamRequestWrittenMs)
	require.Equal(t, http.StatusCreated, *snapshot.FinalUpstreamStatus)
	require.Nil(t, snapshot.AttemptTimeline)
}

func TestAttachHTTPUpstreamDiagnosticsWithoutCollectorIsNoOp(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)

	wrapped, attempt := attachHTTPUpstreamDiagnostics(req, "", 1)

	require.Same(t, req, wrapped)
	require.Nil(t, attempt)
}

type zeroReader struct{}

func (*zeroReader) Read([]byte) (int, error) { return 0, io.EOF }
