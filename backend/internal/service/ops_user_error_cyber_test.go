package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapUserErrorCategoryCyber(t *testing.T) {
	require.Equal(t, "cyber", MapUserErrorCategory("request", "cyber_policy"))
	require.Equal(t, "cyber", MapUserErrorCategory("request", "cyber_policy_session_blocked"))
	phases, types := CategoryToFilter("cyber")
	require.Equal(t, []string{"request"}, phases)
	require.Equal(t, []string{"cyber_policy", "cyber_policy_session_blocked"}, types)
}
