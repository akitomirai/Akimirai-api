package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestV151ImageGenerationIntentNamespaceToolChoice(t *testing.T) {
	require.True(t, IsImageGenerationIntent(
		"/v1/responses",
		"gpt-5.6",
		[]byte(`{"model":"gpt-5.6","tool_choice":{"type":"namespace","name":"image_gen"}}`),
	))
	require.False(t, IsImageGenerationIntent(
		"/v1/responses",
		"gpt-5.6",
		[]byte(`{"model":"gpt-5.6","tool_choice":{"function":{"name":"imagegen"}}}`),
	))
}

func TestV151StripOpenAIImageGenerationNamespaceDeclarations(t *testing.T) {
	reqBody := map[string]any{
		"tools": []any{
			map[string]any{"type": "namespace", "name": "image_gen"},
			map[string]any{"type": "namespace", "name": "code_tools"},
		},
		"input": []any{
			map[string]any{"type": "message", "role": "user", "content": "hello"},
			map[string]any{"type": "additional_tools", "tools": []any{map[string]any{"type": "namespace", "name": "image_gen"}}},
		},
		"tool_choice": map[string]any{"type": "namespace", "name": "image_gen"},
	}

	require.True(t, stripOpenAIImageGenerationTools(reqBody))
	require.False(t, hasOpenAIImageGenerationTool(reqBody))
	require.NotContains(t, reqBody, "tool_choice")
	require.Len(t, reqBody["tools"], 1)
	require.Len(t, reqBody["input"], 1)
}

func TestV151StripOpenAIImageGenerationNamespaceFromRawPayload(t *testing.T) {
	payload := []byte(`{
		"type":"response.create",
		"model":"gpt-5.6",
		"tools":[
			{"type":"namespace","name":"image_gen"},
			{"type":"namespace","name":"code_tools"}
		],
		"input":[
			{"type":"message","role":"user","content":"hello"},
			{"type":"additional_tools","tools":[{"type":"namespace","name":"image_gen"}]}
		],
		"tool_choice":{"type":"namespace","name":"image_gen"}
	}`)

	updated, changed, err := stripOpenAIImageGenerationToolsFromRawPayload(payload)
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, IsImageGenerationIntent(openAIResponsesEndpoint, "gpt-5.6", updated))
	require.True(t, gjson.GetBytes(updated, `tools.#(name=="code_tools")`).Exists())
	require.Equal(t, "hello", gjson.GetBytes(updated, "input.0.content").String())
	require.False(t, gjson.GetBytes(updated, "tool_choice").Exists())
}
