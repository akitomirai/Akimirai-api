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
