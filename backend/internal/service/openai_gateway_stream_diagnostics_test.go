package service

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAIStreamPayloadHasOutputCharacter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload string
		want    bool
	}{
		{name: "preamble", payload: `{"type":"response.created"}`, want: false},
		{name: "empty output delta", payload: `{"type":"response.output_text.delta","delta":""}`, want: false},
		{name: "text output", payload: `{"type":"response.output_text.delta","delta":"x"}`, want: true},
		{name: "leading whitespace is output", payload: `{"type":"response.output_text.delta","delta":" "}`, want: true},
		{name: "function arguments", payload: `{"type":"response.function_call_arguments.delta","delta":"{"}`, want: true},
		{name: "chat role only", payload: `{"choices":[{"delta":{"role":"assistant"}}]}`, want: false},
		{name: "chat content", payload: `{"choices":[{"delta":{"content":"hello"}}]}`, want: true},
		{name: "chat tool arguments", payload: `{"choices":[{"delta":{"tool_calls":[{"function":{"arguments":"{"}}]}}]}`, want: true},
		{name: "usage only", payload: `{"choices":[],"usage":{"prompt_tokens":1}}`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, openAIStreamPayloadHasOutputCharacter([]byte(tt.payload)))
		})
	}
}

func TestMarkOpenAIStreamDiagnosticEventSeparatesEventAndCharacter(t *testing.T) {
	startedAt := time.Now().Add(-time.Second)
	diagnostics := NewRequestDiagnostics(startedAt)
	attempt := diagnostics.BeginHTTPAttemptAt(1, "", startedAt.Add(10*time.Millisecond))
	attempt.FinishAt(startedAt.Add(20*time.Millisecond), http.StatusOK, nil)
	req, err := http.NewRequestWithContext(WithRequestDiagnosticsAttempt(t.Context(), attempt), http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	resp := &http.Response{Request: req}

	markOpenAIStreamDiagnosticEvent(resp, []byte(`{"type":"response.created"}`))
	first := diagnostics.SnapshotAt(time.Now(), nil)
	require.NotNil(t, first.UpstreamFirstEventMs)
	require.Nil(t, first.RequestFirstOutputCharacterMs)

	markOpenAIStreamDiagnosticEvent(resp, []byte(`{"type":"response.output_text.delta","delta":"hello"}`))
	second := diagnostics.SnapshotAt(time.Now(), nil)
	require.Equal(t, first.UpstreamFirstEventMs, second.UpstreamFirstEventMs)
	require.NotNil(t, second.RequestFirstOutputCharacterMs)
}
