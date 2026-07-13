package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeUsageDashboardGranularity(t *testing.T) {
	for _, value := range []string{"hour", "2h", "4h", "8h", "day"} {
		require.Equal(t, value, normalizeUsageDashboardGranularity(value))
	}
	require.Equal(t, "day", normalizeUsageDashboardGranularity("minute"))
	require.Equal(t, "day", normalizeUsageDashboardGranularity("8h; DROP TABLE usage_logs"))
}
