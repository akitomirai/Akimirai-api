package repository

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func attachHTTPUpstreamDiagnostics(req *http.Request, proxyURL string, accountID int64) (*http.Request, *service.RequestDiagnosticsAttempt) {
	if req == nil {
		return nil, nil
	}
	diagnostics := service.RequestDiagnosticsFromContext(req.Context())
	if diagnostics == nil {
		return req, nil
	}

	attempt := diagnostics.BeginHTTPAttemptAt(accountID, proxyURL, time.Now())
	trace := &httptrace.ClientTrace{
		GetConn: func(_ string) {
			attempt.BeginConnectionAcquisitionAt(time.Now())
		},
		GotConn: func(info httptrace.GotConnInfo) {
			attempt.MarkConnectionReadyAt(time.Now(), info.Reused)
		},
		DNSStart: func(_ httptrace.DNSStartInfo) {
			attempt.MarkDNSStartAt(time.Now())
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			attempt.MarkDNSDoneAt(time.Now(), info.Err)
		},
		ConnectStart: func(_, _ string) {
			attempt.MarkConnectStartAt(time.Now())
		},
		ConnectDone: func(_, _ string, err error) {
			attempt.MarkConnectDoneAt(time.Now(), err)
		},
		TLSHandshakeStart: func() {
			attempt.MarkTLSHandshakeStartAt(time.Now())
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			attempt.MarkTLSHandshakeDoneAt(time.Now(), err)
		},
		WroteHeaders: func() {
			attempt.MarkRequestHeadersWrittenAt(time.Now())
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if info.Err == nil {
				attempt.MarkRequestWrittenAt(time.Now())
			}
		},
		GotFirstResponseByte: func() {
			attempt.MarkFirstResponseByteAt(time.Now())
		},
	}
	ctx := service.WithRequestDiagnosticsAttempt(req.Context(), attempt)
	return req.WithContext(httptrace.WithClientTrace(ctx, trace)), attempt
}

func finishHTTPUpstreamDiagnostics(attempt *service.RequestDiagnosticsAttempt, resp *http.Response, err error) {
	if attempt == nil {
		return
	}
	status := 0
	if resp != nil {
		status = resp.StatusCode
		attempt.MarkResponseHeadersReceivedAt(time.Now())
	}
	attempt.FinishAt(time.Now(), status, err)
}

func wrapHTTPUpstreamDiagnosticsBody(resp *http.Response, attempt *service.RequestDiagnosticsAttempt) {
	if resp == nil || resp.Body == nil || attempt == nil {
		return
	}
	resp.Body = &diagnosticsFirstByteBody{ReadCloser: resp.Body, attempt: attempt}
}

type diagnosticsFirstByteBody struct {
	io.ReadCloser
	attempt *service.RequestDiagnosticsAttempt
	once    sync.Once
}

func (b *diagnosticsFirstByteBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		b.once.Do(func() { b.attempt.MarkResponseBodyFirstByteAt(time.Now()) })
	}
	return n, err
}
