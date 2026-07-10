package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/stretchr/testify/require"
)

func TestV151NestedCacheUsageExplicitZeroWins(t *testing.T) {
	usage, ok := extractOpenAIUsageFromJSONBytes([]byte(`{"usage":{"input_tokens":20,"output_tokens":2,"cache_creation_input_tokens":19,"input_tokens_details":{"cache_write_tokens":0,"cached_tokens":0},"cache_read_input_tokens":11}}`))
	require.True(t, ok)
	require.Zero(t, usage.CacheCreationInputTokens)
	require.Zero(t, usage.CacheReadInputTokens)

	canonical := copyOpenAIUsageFromResponsesUsage(&apicompat.ResponsesUsage{
		InputTokens:              20,
		OutputTokens:             2,
		CacheCreationInputTokens: 0,
		InputTokensDetails:       &apicompat.ResponsesInputTokensDetails{CachedTokens: 3, CacheWriteTokens: 19},
	})
	require.Zero(t, canonical.CacheCreationInputTokens)
	require.Equal(t, 3, canonical.CacheReadInputTokens)
}

func TestV151BareGPT56RoutesAndPricesAsSol(t *testing.T) {
	for input, expected := range map[string]string{
		"gpt-5.6":            "gpt-5.6-sol",
		"openai/gpt-5.6":     "gpt-5.6-sol",
		"gpt-5.6-high":       "gpt-5.6-sol",
		"gpt-5.6-max":        "gpt-5.6-sol",
		"gpt-5.6-2026-07-09": "gpt-5.6-sol",
	} {
		require.Equal(t, expected, normalizeKnownOpenAICodexModel(input), input)
	}

	svc := NewBillingService(&config.Config{}, nil)
	tokens := UsageTokens{InputTokens: 100000, CacheCreationTokens: 100000, CacheReadTokens: 73000, OutputTokens: 10}
	cost, err := svc.CalculateCost("gpt-5.6", tokens, 1)
	require.NoError(t, err)
	require.InDelta(t, 100000*5e-6*2, cost.InputCost, 1e-12)
	require.InDelta(t, 100000*6.25e-6*2, cost.CacheCreationCost, 1e-12)
	require.InDelta(t, 73000*0.5e-6*2, cost.CacheReadCost, 1e-12)
	require.InDelta(t, 10*30e-6*1.5, cost.OutputCost, 1e-12)
}
