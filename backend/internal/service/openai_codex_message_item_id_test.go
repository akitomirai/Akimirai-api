//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterCodexInput_StripsMessageItemID_WhenPreservingReferences(t *testing.T) {
	input := []any{map[string]any{
		"type": "message", "id": "item_3bc5a3fa8ccde25f1c0000d4", "role": "user",
		"content": []any{map[string]any{"type": "input_text", "text": "hello"}},
	}}

	filtered := filterCodexInputWithOptions(input, codexInputFilterOptions{PreserveReferences: true})

	require.Len(t, filtered, 1)
	msg, ok := filtered[0].(map[string]any)
	require.True(t, ok)
	_, hasID := msg["id"]
	require.False(t, hasID)
	require.Equal(t, "user", msg["role"])
	require.NotNil(t, msg["content"])
}

func TestFilterCodexInput_KeepsMsgID_WhenPreservingReferences(t *testing.T) {
	input := []any{map[string]any{"type": "message", "id": "msg_validID123", "role": "assistant"}}

	filtered := filterCodexInputWithOptions(input, codexInputFilterOptions{PreserveReferences: true})

	require.Len(t, filtered, 1)
	msg, ok := filtered[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "msg_validID123", msg["id"])
}

func TestFilterCodexInput_StripsMessageIDWhenNotPreservingReferences(t *testing.T) {
	for _, id := range []string{"item_abc", "msg_validID123"} {
		filtered := filterCodexInputWithOptions([]any{map[string]any{
			"type": "message", "id": id, "role": "user",
		}}, codexInputFilterOptions{PreserveReferences: false})

		require.Len(t, filtered, 1)
		msg, ok := filtered[0].(map[string]any)
		require.True(t, ok)
		_, hasID := msg["id"]
		require.False(t, hasID, "id %q should be stripped", id)
	}
}

func TestFilterCodexInput_MessageIDStripDoesNotMutateInput(t *testing.T) {
	original := map[string]any{"type": "message", "id": "item_abc", "role": "user"}

	filtered := filterCodexInputWithOptions([]any{original}, codexInputFilterOptions{PreserveReferences: true})

	require.Len(t, filtered, 1)
	require.Equal(t, "item_abc", original["id"])
}

func TestFilterCodexInput_MessageStripKeepsFunctionCallBehavior(t *testing.T) {
	input := []any{
		map[string]any{"type": "message", "id": "item_msg_001", "role": "user"},
		map[string]any{"type": "function_call", "id": "fc_validID123", "call_id": "fc_validID123", "name": "bash"},
		map[string]any{"type": "function_call", "id": "item_bad", "call_id": "fc_abc123", "name": "bash"},
		map[string]any{"type": "function_call_output", "id": "o1", "call_id": "fc_abc123", "output": "done"},
	}

	filtered := filterCodexInputWithOptions(input, codexInputFilterOptions{PreserveReferences: true})

	require.Len(t, filtered, 4)
	msg := filtered[0].(map[string]any)
	_, hasID := msg["id"]
	require.False(t, hasID)
	require.Equal(t, "fc_validID123", filtered[1].(map[string]any)["id"])
	_, hasID = filtered[2].(map[string]any)["id"]
	require.False(t, hasID)
	require.Equal(t, "fc_abc123", filtered[2].(map[string]any)["call_id"])
	require.Equal(t, "o1", filtered[3].(map[string]any)["id"])
	require.Equal(t, "fc_abc123", filtered[3].(map[string]any)["call_id"])
}
