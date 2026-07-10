package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV151ResponsesUsageNestedCacheWritePresenceOverridesTopLevel(t *testing.T) {
	for _, tt := range []struct {
		name, nested string
		want         int
	}{
		{name: "explicit zero", nested: `{"cache_write_tokens":0}`, want: 0},
		{name: "nonzero", nested: `{"cache_write_tokens":7}`, want: 7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var usage ResponsesUsage
			payload := []byte(`{"input_tokens":20,"output_tokens":2,"cache_creation_input_tokens":19,"input_tokens_details":` + tt.nested + `}`)
			require.NoError(t, json.Unmarshal(payload, &usage))
			require.Equal(t, tt.want, usage.CacheCreationInputTokens)
		})
	}
}
