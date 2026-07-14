package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

func markOpenAIStreamDiagnosticEvent(resp *http.Response, payload []byte) {
	if resp == nil || resp.Request == nil || len(payload) == 0 {
		return
	}
	trimmed := strings.TrimSpace(string(payload))
	if trimmed == "" || trimmed == "[DONE]" {
		return
	}
	attempt := RequestDiagnosticsAttemptFromContext(resp.Request.Context())
	if attempt == nil {
		return
	}
	now := time.Now()
	attempt.MarkFirstStreamEventAt(now)
	if openAIStreamPayloadHasOutputCharacter(payload) {
		attempt.MarkFirstOutputCharacterAt(now)
	}
}

func openAIStreamPayloadHasOutputCharacter(payload []byte) bool {
	if len(payload) == 0 || !gjson.ValidBytes(payload) {
		return false
	}

	eventType := strings.TrimSpace(gjson.GetBytes(payload, "type").String())
	switch eventType {
	case "response.output_text.delta",
		"response.reasoning_text.delta",
		"response.reasoning_summary_text.delta",
		"response.function_call_arguments.delta",
		"response.mcp_call_arguments.delta",
		"response.custom_tool_call_input.delta":
		return gjson.GetBytes(payload, "delta").String() != ""
	case "content_block_delta":
		return firstNonEmptyJSONText(payload,
			"delta.text",
			"delta.partial_json",
			"delta.thinking",
		)
	}

	for _, choice := range gjson.GetBytes(payload, "choices").Array() {
		if firstNonEmptyResultText(choice,
			"text",
			"delta.content",
			"delta.reasoning_content",
			"delta.reasoning",
		) {
			return true
		}
		for _, toolCall := range choice.Get("delta.tool_calls").Array() {
			if toolCall.Get("function.arguments").String() != "" {
				return true
			}
		}
	}
	return false
}

func firstNonEmptyJSONText(payload []byte, paths ...string) bool {
	for _, path := range paths {
		if gjson.GetBytes(payload, path).String() != "" {
			return true
		}
	}
	return false
}

func firstNonEmptyResultText(value gjson.Result, paths ...string) bool {
	for _, path := range paths {
		if value.Get(path).String() != "" {
			return true
		}
	}
	return false
}
