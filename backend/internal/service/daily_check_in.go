package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"
)

const dailyCheckInResetHour = 2

var dailyCheckInLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

// DailyCheckInRecord is the immutable balance grant persisted for one service day.
type DailyCheckInRecord struct {
	ID            int64
	UserID        int64
	ServiceDate   string
	RewardAmount  float64
	BalanceBefore float64
	BalanceAfter  float64
	CheckedInAt   time.Time
	CreatedAt     time.Time
}

// DailyCheckInClaim contains the transaction inputs owned by the repository.
type DailyCheckInClaim struct {
	UserID       int64
	ServiceDate  string
	RewardAmount float64
	CheckedInAt  time.Time
}

// DailyCheckInRepository owns exact-once persistence and the balance mutation.
type DailyCheckInRepository interface {
	GetForServiceDate(ctx context.Context, userID int64, serviceDate string) (*DailyCheckInRecord, float64, error)
	Claim(ctx context.Context, claim DailyCheckInClaim) (*DailyCheckInRecord, bool, error)
}

// DailyCheckInStatus is the stable API contract used by both GET and POST.
type DailyCheckInStatus struct {
	CheckedIn        bool       `json:"checked_in"`
	AlreadyCheckedIn bool       `json:"already_checked_in"`
	ServiceDate      string     `json:"service_date"`
	RewardAmount     float64    `json:"reward_amount"`
	BalanceBefore    float64    `json:"balance_before"`
	BalanceAfter     float64    `json:"balance_after"`
	CheckedInAt      *time.Time `json:"checked_in_at"`
	NextResetAt      time.Time  `json:"next_reset_at"`
}

type dailyCheckInBalanceInvalidator interface {
	InvalidateUserBalance(ctx context.Context, userID int64) error
}

// DailyCheckInService owns service-day math, reward generation, and cache coherence.
type DailyCheckInService struct {
	repo               DailyCheckInRepository
	authInvalidator    APIKeyAuthCacheInvalidator
	balanceInvalidator dailyCheckInBalanceInvalidator
	now                func() time.Time
	reward             func() (int, error)
}

func NewDailyCheckInService(
	repo DailyCheckInRepository,
	authInvalidator APIKeyAuthCacheInvalidator,
	billingCache BillingCache,
) *DailyCheckInService {
	return newDailyCheckInService(repo, authInvalidator, billingCache)
}

func newDailyCheckInService(
	repo DailyCheckInRepository,
	authInvalidator APIKeyAuthCacheInvalidator,
	balanceInvalidator dailyCheckInBalanceInvalidator,
) *DailyCheckInService {
	return &DailyCheckInService{
		repo:               repo,
		authInvalidator:    authInvalidator,
		balanceInvalidator: balanceInvalidator,
		now:                time.Now,
		reward:             secureDailyCheckInReward,
	}
}

func secureDailyCheckInReward() (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(3))
	if err != nil {
		return 0, fmt.Errorf("generate daily check-in reward: %w", err)
	}
	return int(value.Int64()) + 1, nil
}

func dailyCheckInWindow(now time.Time) (string, time.Time) {
	localNow := now.In(dailyCheckInLocation)
	shifted := localNow.Add(-dailyCheckInResetHour * time.Hour)
	year, month, day := shifted.Date()
	serviceDate := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	nextResetAt := time.Date(year, month, day+1, dailyCheckInResetHour, 0, 0, 0, dailyCheckInLocation)
	return serviceDate, nextResetAt
}

func (s *DailyCheckInService) GetStatus(ctx context.Context, userID int64) (*DailyCheckInStatus, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("daily check-in repository is unavailable")
	}

	serviceDate, nextResetAt := dailyCheckInWindow(s.now())
	record, currentBalance, err := s.repo.GetForServiceDate(ctx, userID, serviceDate)
	if err != nil {
		return nil, fmt.Errorf("get daily check-in status: %w", err)
	}
	if record == nil {
		return &DailyCheckInStatus{
			CheckedIn:        false,
			AlreadyCheckedIn: false,
			ServiceDate:      serviceDate,
			RewardAmount:     0,
			BalanceBefore:    currentBalance,
			BalanceAfter:     currentBalance,
			CheckedInAt:      nil,
			NextResetAt:      nextResetAt,
		}, nil
	}
	return dailyCheckInStatusFromRecord(record, true, nextResetAt), nil
}

func (s *DailyCheckInService) Claim(ctx context.Context, userID int64) (*DailyCheckInStatus, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("daily check-in repository is unavailable")
	}

	now := s.now()
	serviceDate, nextResetAt := dailyCheckInWindow(now)
	reward, err := s.reward()
	if err != nil {
		return nil, err
	}
	if reward < 1 || reward > 3 {
		return nil, fmt.Errorf("daily check-in reward must be between 1 and 3")
	}

	record, created, err := s.repo.Claim(ctx, DailyCheckInClaim{
		UserID:       userID,
		ServiceDate:  serviceDate,
		RewardAmount: float64(reward),
		CheckedInAt:  now,
	})
	if err != nil {
		return nil, fmt.Errorf("claim daily check-in reward: %w", err)
	}
	if record == nil {
		return nil, fmt.Errorf("claim daily check-in reward: repository returned no record")
	}

	if created {
		s.invalidateBalanceCaches(ctx, userID)
	}
	return dailyCheckInStatusFromRecord(record, !created, nextResetAt), nil
}

func dailyCheckInStatusFromRecord(record *DailyCheckInRecord, alreadyCheckedIn bool, nextResetAt time.Time) *DailyCheckInStatus {
	checkedInAt := record.CheckedInAt
	return &DailyCheckInStatus{
		CheckedIn:        true,
		AlreadyCheckedIn: alreadyCheckedIn,
		ServiceDate:      record.ServiceDate,
		RewardAmount:     record.RewardAmount,
		BalanceBefore:    record.BalanceBefore,
		BalanceAfter:     record.BalanceAfter,
		CheckedInAt:      &checkedInAt,
		NextResetAt:      nextResetAt,
	}
}

func (s *DailyCheckInService) invalidateBalanceCaches(ctx context.Context, userID int64) {
	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if s.authInvalidator != nil {
		s.authInvalidator.InvalidateAuthCacheByUserID(cacheCtx, userID)
	}
	if s.balanceInvalidator != nil {
		if err := s.balanceInvalidator.InvalidateUserBalance(cacheCtx, userID); err != nil {
			slog.Warn("invalidate daily check-in balance cache", "user_id", userID, "error", err)
		}
	}
}
