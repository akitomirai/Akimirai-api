package handler

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func beginOpenAIRequestDiagnostics(c *gin.Context, startedAt time.Time) *service.RequestDiagnostics {
	diagnostics := service.NewRequestDiagnostics(startedAt)
	if c == nil || c.Request == nil {
		return diagnostics
	}
	c.Request = c.Request.WithContext(service.WithRequestDiagnostics(c.Request.Context(), diagnostics))
	return diagnostics
}

func (h *OpenAIGatewayHandler) readOpenAIRequestBodyWithDiagnostics(c *gin.Context, diagnostics *service.RequestDiagnostics) ([]byte, error) {
	startedAt := time.Now()
	body, err := readLenientJSONRequestBodyWithPrealloc(c.Request, h.cfg)
	if diagnostics != nil {
		diagnostics.RecordBodyRead(time.Since(startedAt), int64(len(body)))
		if len(body) > 0 {
			effectiveKey := ""
			if h != nil && h.gatewayService != nil {
				effectiveKey = h.gatewayService.ExtractSessionID(c, body)
			}
			diagnostics.RecordPromptCacheDiagnostics(service.PromptCacheDiagnosticsForRequest(c, body, effectiveKey, false))
		}
	}
	return body, err
}

func snapshotOpenAIRequestDiagnostics(
	diagnostics *service.RequestDiagnostics,
	forwardStartedAt time.Time,
	result *service.OpenAIForwardResult,
	account *service.Account,
) *service.RequestDiagnosticsSnapshot {
	if diagnostics == nil {
		return nil
	}
	if result != nil {
		diagnostics.MarkFirstSemanticToken(forwardStartedAt, result.FirstTokenMs)
	}
	snapshot := diagnostics.SnapshotAt(time.Now(), account)
	return &snapshot
}
