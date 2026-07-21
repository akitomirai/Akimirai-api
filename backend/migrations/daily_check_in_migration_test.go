package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDailyCheckInMigrationOwnsExactOnceAndBalanceInvariants(t *testing.T) {
	content, err := FS.ReadFile("192_daily_check_ins.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "UNIQUE (user_id, service_date)")
	require.Contains(t, sql, "reward_amount DECIMAL(20, 8)")
	require.Contains(t, sql, "reward_amount IN (1, 2, 3)")
	require.Contains(t, sql, "balance_after = balance_before + reward_amount")
	require.Contains(t, sql, "REFERENCES users(id) ON DELETE CASCADE")
	require.Contains(t, sql, "checked_in_at TIMESTAMPTZ NOT NULL")
}
