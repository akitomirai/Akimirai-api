package apicompat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicUsageFromResponsesUsage_CacheCreation(t *testing.T) {
	usage := &ResponsesUsage{
		InputTokens:              20,
		OutputTokens:             5,
		CacheCreationInputTokens: 6,
		InputTokensDetails:       &ResponsesInputTokensDetails{CachedTokens: 4},
	}

	got := anthropicUsageFromResponsesUsage(usage)

	assert.Equal(t, 10, got.InputTokens)
	assert.Equal(t, 5, got.OutputTokens)
	assert.Equal(t, 4, got.CacheReadInputTokens)
	assert.Equal(t, 6, got.CacheCreationInputTokens)
}

func TestAnthropicUsageFromResponsesUsage_NoCacheCreation(t *testing.T) {
	usage := &ResponsesUsage{
		InputTokens:        10,
		OutputTokens:       5,
		InputTokensDetails: &ResponsesInputTokensDetails{CachedTokens: 3},
	}

	got := anthropicUsageFromResponsesUsage(usage)

	assert.Equal(t, 7, got.InputTokens)
	assert.Equal(t, 3, got.CacheReadInputTokens)
	assert.Zero(t, got.CacheCreationInputTokens)
}

func TestResponsesEventToAnthropicEvents_StreamingCacheCreation(t *testing.T) {
	state := NewResponsesEventToAnthropicState()
	state.MessageStartSent = true
	completedEvt := &ResponsesStreamEvent{
		Type: "response.completed",
		Response: &ResponsesResponse{
			Status: "completed",
			Usage: &ResponsesUsage{
				InputTokens:              20,
				OutputTokens:             5,
				CacheCreationInputTokens: 6,
				InputTokensDetails:       &ResponsesInputTokensDetails{CachedTokens: 4},
			},
		},
	}

	events := ResponsesEventToAnthropicEvents(completedEvt, state)
	var deltaEvt *AnthropicStreamEvent
	for i := range events {
		if events[i].Type == "message_delta" {
			deltaEvt = &events[i]
			break
		}
	}

	require.NotNil(t, deltaEvt)
	require.NotNil(t, deltaEvt.Usage)
	assert.Equal(t, 6, deltaEvt.Usage.CacheCreationInputTokens)
	assert.Equal(t, 10, deltaEvt.Usage.InputTokens)
	assert.Equal(t, 4, deltaEvt.Usage.CacheReadInputTokens)
}

func TestAnthropicToResponsesResponse_CacheCreation(t *testing.T) {
	resp := AnthropicResponse{
		ID: "msg_test", Type: "message", Role: "assistant", Model: "claude-opus-4-6",
		Usage: AnthropicUsage{
			InputTokens: 10, OutputTokens: 5, CacheReadInputTokens: 4, CacheCreationInputTokens: 6,
		},
		StopReason: AnthropicStopReasonPtr("end_turn"),
	}

	out := AnthropicToResponsesResponse(&resp)

	require.NotNil(t, out.Usage)
	assert.Equal(t, 20, out.Usage.InputTokens)
	assert.Equal(t, 6, out.Usage.CacheCreationInputTokens)
}
