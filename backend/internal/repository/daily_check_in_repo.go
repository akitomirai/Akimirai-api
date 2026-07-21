package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type dailyCheckInRepository struct {
	db *sql.DB
}

func NewDailyCheckInRepository(db *sql.DB) service.DailyCheckInRepository {
	return &dailyCheckInRepository{db: db}
}

func (r *dailyCheckInRepository) GetForServiceDate(
	ctx context.Context,
	userID int64,
	serviceDate string,
) (*service.DailyCheckInRecord, float64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("nil daily check-in database")
	}

	var balance float64
	err := r.db.QueryRowContext(ctx, `
		SELECT balance::double precision FROM users
		WHERE id = $1 AND status = 'active' AND deleted_at IS NULL
	`, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, service.ErrUserNotFound
	}
	if err != nil {
		return nil, 0, fmt.Errorf("load daily check-in user balance: %w", err)
	}

	record, err := selectDailyCheckIn(ctx, r.db, userID, serviceDate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, balance, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("load daily check-in record: %w", err)
	}
	return record, balance, nil
}

func (r *dailyCheckInRepository) Claim(
	ctx context.Context,
	claim service.DailyCheckInClaim,
) (*service.DailyCheckInRecord, bool, error) {
	if r == nil || r.db == nil {
		return nil, false, errors.New("nil daily check-in database")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, fmt.Errorf("begin daily check-in transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var balanceBefore float64
	err = tx.QueryRowContext(ctx, `
		SELECT balance::double precision FROM users
		WHERE id = $1 AND status = 'active' AND deleted_at IS NULL
		FOR UPDATE
	`, claim.UserID).Scan(&balanceBefore)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, service.ErrUserNotFound
	}
	if err != nil {
		return nil, false, fmt.Errorf("lock daily check-in user: %w", err)
	}

	balanceAfter := balanceBefore + claim.RewardAmount
	record, err := scanDailyCheckIn(tx.QueryRowContext(ctx, `
		INSERT INTO daily_check_ins (
			user_id, service_date, reward_amount,
			balance_before, balance_after, checked_in_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, service_date) DO NOTHING
		RETURNING id, user_id, service_date::text, reward_amount::double precision,
			balance_before::double precision, balance_after::double precision,
			checked_in_at, created_at
	`, claim.UserID, claim.ServiceDate, claim.RewardAmount, balanceBefore, balanceAfter, claim.CheckedInAt))
	if errors.Is(err, sql.ErrNoRows) {
		record, err = selectDailyCheckIn(ctx, tx, claim.UserID, claim.ServiceDate)
		if err != nil {
			return nil, false, fmt.Errorf("load existing daily check-in record: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, false, fmt.Errorf("commit replayed daily check-in transaction: %w", err)
		}
		return record, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("insert daily check-in record: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE users SET balance = $2, updated_at = $3
		WHERE id = $1 AND status = 'active' AND deleted_at IS NULL
	`, claim.UserID, balanceAfter, claim.CheckedInAt)
	if err != nil {
		return nil, false, fmt.Errorf("apply daily check-in balance: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("inspect daily check-in balance update: %w", err)
	}
	if affected != 1 {
		return nil, false, fmt.Errorf("apply daily check-in balance: expected one user, updated %d", affected)
	}

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit daily check-in transaction: %w", err)
	}
	return record, true, nil
}

type dailyCheckInQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func selectDailyCheckIn(
	ctx context.Context,
	queryer dailyCheckInQueryer,
	userID int64,
	serviceDate string,
) (*service.DailyCheckInRecord, error) {
	return scanDailyCheckIn(queryer.QueryRowContext(ctx, `
		SELECT id, user_id, service_date::text, reward_amount::double precision,
			balance_before::double precision, balance_after::double precision,
			checked_in_at, created_at
		FROM daily_check_ins
		WHERE user_id = $1 AND service_date = $2
	`, userID, serviceDate))
}

type dailyCheckInScanner interface {
	Scan(dest ...any) error
}

func scanDailyCheckIn(row dailyCheckInScanner) (*service.DailyCheckInRecord, error) {
	record := &service.DailyCheckInRecord{}
	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.ServiceDate,
		&record.RewardAmount,
		&record.BalanceBefore,
		&record.BalanceAfter,
		&record.CheckedInAt,
		&record.CreatedAt,
	); err != nil {
		return nil, err
	}
	return record, nil
}
