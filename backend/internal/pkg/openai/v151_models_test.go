package openai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV151DefaultModelsIncludeBareGPT56Alias(t *testing.T) {
	require.Contains(t, DefaultModelIDs(), "gpt-5.6")
}
