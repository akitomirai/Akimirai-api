package openai_ws_v2

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestV151NestedCacheWriteZeroWins(t *testing.T) {
	usage := gjson.Parse(`{"input_tokens_details":{"cache_write_tokens":0},"cache_creation_input_tokens":19}`)
	require.Zero(t, openAICacheCreationTokensFromUsage(usage))
}
