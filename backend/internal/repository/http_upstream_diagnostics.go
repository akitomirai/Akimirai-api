package repository

import (
	"net/http"
	"net/http/httptrace"
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
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if info.Err == nil {
				attempt.MarkRequestWrittenAt(time.Now())
			}
		},
		GotFirstResponseByte: func() {
			attempt.MarkFirstResponseByteAt(time.Now())
		},
	}
	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace)), attempt
}

func finishHTTPUpstreamDiagnostics(attempt *service.RequestDiagnosticsAttempt, resp *http.Response, err error) {
	if attempt == nil {
		return
	}
	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	attempt.FinishAt(time.Now(), status, err)
}
