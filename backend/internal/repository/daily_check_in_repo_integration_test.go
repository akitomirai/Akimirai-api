//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDailyCheckInRepositoryConcurrentClaimsCreditExactlyOnce(t *testing.T) {
	ctx := context.Background()
	user := mustCreateUser(t, integrationEntClient, &service.User{
		Email:   fmt.Sprintf("daily-check-in-%d@example.com", time.Now().UnixNano()),
		Balance: 10,
	})
	_, err := integrationDB.ExecContext(ctx, `UPDATE users SET total_recharged = 7 WHERE id = $1`, user.ID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
	})

	repo := NewDailyCheckInRepository(integrationDB)
	checkedAt := time.Date(2026, 7, 21, 3, 4, 5, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	const callers = 16
	var createdCount atomic.Int64
	records := make(chan *service.DailyCheckInRecord, callers)
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			record, created, claimErr := repo.Claim(ctx, service.DailyCheckInClaim{
				UserID: user.ID, ServiceDate: "2026-07-21", RewardAmount: 2, CheckedInAt: checkedAt,
			})
			if claimErr != nil {
				errs <- claimErr
				return
			}
			if created {
				createdCount.Add(1)
			}
			records <- record
		}()
	}
	wg.Wait()
	close(errs)
	close(records)
	for claimErr := range errs {
		require.NoError(t, claimErr)
	}
	require.Equal(t, int64(1), createdCount.Load())
	for record := range records {
		require.Equal(t, float64(2), record.RewardAmount)
		require.Equal(t, float64(10), record.BalanceBefore)
		require.Equal(t, float64(12), record.BalanceAfter)
	}

	var balance, totalRecharged float64
	var ledgerCount int
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		`SELECT balance::double precision, total_recharged::double precision FROM users WHERE id = $1`, user.ID,
	).Scan(&balance, &totalRecharged))
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM daily_check_ins WHERE user_id = $1 AND service_date = $2`, user.ID, "2026-07-21",
	).Scan(&ledgerCount))
	require.Equal(t, float64(12), balance)
	require.Equal(t, float64(7), totalRecharged)
	require.Equal(t, 1, ledgerCount)
}
