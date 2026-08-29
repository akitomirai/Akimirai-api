package apicompat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRetypedResponsesToolCallItemID(t *testing.T) {
	for _, tc := range []struct {
		name     string
		id       string
		itemType string
		want     string
	}{
		{"function id raised to custom", "fc_abc", "custom_tool_call", "ctc_abc"},
		{"function id raised to tool search", "fc_abc", "tool_search_call", "tsc_abc"},
		{"already correct is untouched", "ctc_abc", "custom_tool_call", "ctc_abc"},
		{"custom id lowered to function", "ctc_abc", "function_call", "fc_abc"},
		{"unknown prefix is left alone", "item_abc", "custom_tool_call", "item_abc"},
		{"unprefixed id is left alone", "abc", "custom_tool_call", "abc"},
		{"empty id stays empty", "", "custom_tool_call", ""},
		{"unconstrained item type is left alone", "fc_abc", "message", "fc_abc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, retypedResponsesToolCallItemID(tc.id, tc.itemType))
		})
	}
}

func TestRestoreResponsesClientToolPayload_RetypesToolCallItemIDs(t *testing.T) {
	mapping := ResponsesClientToolMapping{
		CustomTools:    map[string]bool{"exec": true},
		ToolSearch:     true,
		NamespaceTools: map[string]ResponsesNamespaceName{"team__send": {Namespace: "team", Name: "send"}},
	}
	payload := []byte(`{"id":"resp","output":[` +
		`{"type":"function_call","id":"fc_abc123","call_id":"call_1","name":"exec","arguments":"{\"input\":\"dir\"}"},` +
		`{"type":"function_call","id":"fc_def456","call_id":"call_2","name":"tool_search","arguments":"{\"query\":\"git\"}"},` +
		`{"type":"function_call","id":"fc_ghi789","call_id":"call_3","name":"team__send","arguments":"{}"}]}`)

	restored, changed, err := RestoreResponsesClientToolPayload(payload, mapping)
	require.NoError(t, err)
	require.True(t, changed)
	require.JSONEq(t, `{"id":"resp","output":[`+
		`{"type":"custom_tool_call","id":"ctc_abc123","call_id":"call_1","name":"exec","input":"dir"},`+
		`{"type":"tool_search_call","id":"tsc_def456","call_id":"call_2","execution":"client","arguments":{"query":"git"}},`+
		`{"type":"function_call","id":"fc_ghi789","call_id":"call_3","name":"send","namespace":"team","arguments":"{}"}]}`,
		string(restored))
}

func TestRestoreResponsesOutputClientTools_RetypesToolCallItemIDs(t *testing.T) {
	mapping := ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}, ToolSearch: true}
	outputs := []ResponsesOutput{
		{Type: "function_call", ID: "fc_abc123", CallID: "call_1", Name: "exec", Arguments: `{"input":"dir"}`},
		{Type: "function_call", ID: "fc_def456", CallID: "call_2", Name: toolSearchProxyName, Arguments: `{"query":"git"}`},
	}

	restoreResponsesOutputClientTools(outputs, &mapping)

	require.Equal(t, "custom_tool_call", outputs[0].Type)
	require.Equal(t, "ctc_abc123", outputs[0].ID)
	require.Equal(t, "tool_search_call", outputs[1].Type)
	require.Equal(t, "tsc_def456", outputs[1].ID)
}

func TestResponsesClientToolStreamRestorer_RetypesItemIDAndKeepsUpstreamLookup(t *testing.T) {
	const upstreamID = "fc_09f77ac43cf7db36016a8920e7934487"
	const clientID = "ctc_09f77ac43cf7db36016a8920e7934487"

	restorer := NewResponsesClientToolStreamRestorer(ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}})
	added := restorer.Restore(ResponsesStreamEvent{
		Type: "response.output_item.added", SequenceNumber: 0, OutputIndex: 0,
		Item: &ResponsesOutput{Type: "function_call", ID: upstreamID, CallID: "call_1", Name: "exec", Status: "in_progress"},
	})
	require.Len(t, added, 1)
	require.Equal(t, "custom_tool_call", added[0].Item.Type)
	require.Equal(t, clientID, added[0].Item.ID)

	require.Empty(t, restorer.Restore(ResponsesStreamEvent{
		Type: "response.function_call_arguments.delta", SequenceNumber: 1, ItemID: upstreamID, Delta: `{"input":"di`,
	}))
	done := restorer.Restore(ResponsesStreamEvent{
		Type: "response.function_call_arguments.done", SequenceNumber: 2, ItemID: upstreamID,
		CallID: "call_1", Name: "exec", Arguments: `{"input":"dir"}`,
	})
	require.Len(t, done, 2)
	require.Equal(t, clientID, done[0].ItemID)
	require.Equal(t, clientID, done[1].ItemID)

	closed := restorer.Restore(ResponsesStreamEvent{
		Type: "response.output_item.done", SequenceNumber: 3, OutputIndex: 0,
		Item: &ResponsesOutput{Type: "function_call", ID: upstreamID, CallID: "call_1", Name: "exec", Arguments: `{"input":"dir"}`, Status: "completed"},
	})
	require.Len(t, closed, 1)
	require.Equal(t, "custom_tool_call", closed[0].Item.Type)
	require.Equal(t, clientID, closed[0].Item.ID)
	require.Equal(t, "dir", closed[0].Item.Input)
}

func TestAdaptResponsesClientTools_RecoversRetypedToolCallItemID(t *testing.T) {
	req := map[string]any{
		"tools": []any{
			map[string]any{"type": "custom", "name": "exec"},
			map[string]any{"type": "tool_search"},
		},
		"input": []any{
			map[string]any{"type": "custom_tool_call", "id": "ctc_upstream1", "call_id": "call_1", "name": "exec", "input": "dir"},
			map[string]any{"type": "tool_search_call", "id": "tsc_upstream2", "call_id": "call_2", "arguments": map[string]any{"query": "git"}},
			map[string]any{"type": "custom_tool_call_output", "id": "ctco_client", "call_id": "call_1", "output": "ok"},
		},
	}

	_, changed, err := AdaptResponsesClientTools(req)
	require.NoError(t, err)
	require.True(t, changed)
	input := requireResponsesClientToolValue[[]any](t, req["input"])
	customCall := requireResponsesClientToolValue[map[string]any](t, input[0])
	require.Equal(t, "fc_upstream1", customCall["id"])
	searchCall := requireResponsesClientToolValue[map[string]any](t, input[1])
	require.Equal(t, "fc_upstream2", searchCall["id"])
	customOutput := requireResponsesClientToolValue[map[string]any](t, input[2])
	require.NotContains(t, customOutput, "id")
}
