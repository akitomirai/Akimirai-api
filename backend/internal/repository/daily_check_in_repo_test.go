package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDailyCheckInRepositoryClaimCreatesLedgerAndUpdatesOnlyBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	checkedAt := time.Date(2026, 7, 21, 3, 4, 5, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT balance::double precision FROM users.*FOR UPDATE`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(10.0))
	mock.ExpectQuery(`INSERT INTO daily_check_ins`).
		WithArgs(int64(7), "2026-07-21", float64(2), float64(10), float64(12), checkedAt).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "service_date", "reward_amount", "balance_before", "balance_after", "checked_in_at", "created_at",
		}).AddRow(41, 7, "2026-07-21", 2.0, 10.0, 12.0, checkedAt, checkedAt))
	mock.ExpectExec(`UPDATE users SET balance = \$2, updated_at = \$3`).
		WithArgs(int64(7), float64(12), checkedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewDailyCheckInRepository(db)
	record, created, err := repo.Claim(context.Background(), service.DailyCheckInClaim{
		UserID: 7, ServiceDate: "2026-07-21", RewardAmount: 2, CheckedInAt: checkedAt,
	})
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, int64(41), record.ID)
	require.Equal(t, float64(10), record.BalanceBefore)
	require.Equal(t, float64(12), record.BalanceAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDailyCheckInRepositoryClaimReplayReturnsOriginalWithoutBalanceUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	checkedAt := time.Date(2026, 7, 21, 3, 4, 5, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT balance::double precision FROM users.*FOR UPDATE`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(12.0))
	mock.ExpectQuery(`INSERT INTO daily_check_ins`).
		WithArgs(int64(7), "2026-07-21", float64(3), float64(12), float64(15), checkedAt).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT id, user_id, service_date::text`).
		WithArgs(int64(7), "2026-07-21").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "service_date", "reward_amount", "balance_before", "balance_after", "checked_in_at", "created_at",
		}).AddRow(41, 7, "2026-07-21", 1.0, 10.0, 11.0, checkedAt.Add(-time.Hour), checkedAt.Add(-time.Hour)))
	mock.ExpectCommit()

	repo := NewDailyCheckInRepository(db)
	record, created, err := repo.Claim(context.Background(), service.DailyCheckInClaim{
		UserID: 7, ServiceDate: "2026-07-21", RewardAmount: 3, CheckedInAt: checkedAt,
	})
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, float64(1), record.RewardAmount)
	require.Equal(t, float64(11), record.BalanceAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDailyCheckInRepositoryGetForServiceDateReturnsCurrentBalanceWhenAvailable(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(`SELECT balance::double precision FROM users`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(9.5))
	mock.ExpectQuery(`SELECT id, user_id, service_date::text`).
		WithArgs(int64(7), "2026-07-21").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "service_date", "reward_amount", "balance_before", "balance_after", "checked_in_at", "created_at",
		}))

	repo := NewDailyCheckInRepository(db)
	record, balance, err := repo.GetForServiceDate(context.Background(), 7, "2026-07-21")
	require.NoError(t, err)
	require.Nil(t, record)
	require.Equal(t, 9.5, balance)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDailyCheckInRepositoryListForAdminFiltersAndScansLedgerRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	checkedAt := time.Date(2026, 7, 21, 3, 4, 5, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	filter := service.DailyCheckInAdminFilter{
		Page: 2, PageSize: 25, Query: "alice", ServiceDate: "2026-07-21",
	}
	likeQuery := "%alice%"
	mock.ExpectQuery(`SELECT COUNT\(\*\).*FROM daily_check_ins d.*JOIN users u.*d\.user_id::text ILIKE \$1.*d\.service_date = \$2`).
		WithArgs(likeQuery, filter.ServiceDate).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT d\.id, d\.user_id, u\.email, u\.username.*FROM daily_check_ins d.*JOIN users u.*ORDER BY d\.checked_in_at DESC, d\.id DESC.*LIMIT \$3 OFFSET \$4`).
		WithArgs(likeQuery, filter.ServiceDate, filter.PageSize, 25).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "email", "username", "service_date", "reward_amount",
			"balance_before", "balance_after", "checked_in_at", "created_at",
		}).AddRow(41, 7, "alice@example.com", "Alice", "2026-07-21", 2.0, 10.0, 12.0, checkedAt, checkedAt))

	repo := NewDailyCheckInRepository(db)
	items, total, err := repo.ListForAdmin(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, int64(41), items[0].ID)
	require.Equal(t, "alice@example.com", items[0].Email)
	require.Equal(t, float64(2), items[0].RewardAmount)
	require.Equal(t, float64(12), items[0].BalanceAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDailyCheckInRepositoryListForAdminAllDatesUsesStablePagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(`SELECT COUNT\(\*\).*FROM daily_check_ins d.*JOIN users u`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`FROM daily_check_ins d.*JOIN users u.*ORDER BY d\.checked_in_at DESC, d\.id DESC.*LIMIT \$1 OFFSET \$2`).
		WithArgs(20, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "email", "username", "service_date", "reward_amount",
			"balance_before", "balance_after", "checked_in_at", "created_at",
		}))

	repo := NewDailyCheckInRepository(db)
	items, total, err := repo.ListForAdmin(context.Background(), service.DailyCheckInAdminFilter{
		Page: 2, PageSize: 20, AllDates: true,
	})
	require.NoError(t, err)
	require.Empty(t, items)
	require.Zero(t, total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDailyCheckInRepositoryListForUserUsesStablePagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	checkedAt := time.Date(2026, 7, 21, 3, 4, 5, 0, time.UTC)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM daily_check_ins WHERE user_id = \$1`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, user_id, service_date::text, reward_amount::double precision.*FROM daily_check_ins.*WHERE user_id = \$1.*ORDER BY checked_in_at DESC, id DESC.*LIMIT \$2 OFFSET \$3`).
		WithArgs(int64(7), 20, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "service_date", "reward_amount", "balance_before", "balance_after", "checked_in_at", "created_at",
		}).AddRow(41, 7, "2026-07-21", 0.25, 10.0, 10.25, checkedAt, checkedAt))

	repo := NewDailyCheckInRepository(db)
	items, total, err := repo.ListForUser(context.Background(), 7, 2, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, float64(0.25), items[0].RewardAmount)
	require.NoError(t, mock.ExpectationsWereMet())
}
