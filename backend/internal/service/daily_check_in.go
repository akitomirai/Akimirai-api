package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	dailyCheckInResetHour       = 2
	dailyCheckInRewardRollCount = int64(100)
)

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

// DailyCheckInAdminFilter is the normalized read-only administrator query.
type DailyCheckInAdminFilter struct {
	Page        int
	PageSize    int
	Query       string
	ServiceDate string
	AllDates    bool
}

// DailyCheckInAdminRecord combines the immutable ledger row with display identity.
type DailyCheckInAdminRecord struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	ServiceDate   string    `json:"service_date"`
	RewardAmount  float64   `json:"reward_amount"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	CheckedInAt   time.Time `json:"checked_in_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// DailyCheckInAdminList is the service result consumed by the admin handler.
type DailyCheckInAdminList struct {
	Items    []DailyCheckInAdminRecord
	Total    int64
	Page     int
	PageSize int
}

// DailyCheckInRepository owns exact-once persistence and the balance mutation.
type DailyCheckInRepository interface {
	GetForServiceDate(ctx context.Context, userID int64, serviceDate string) (*DailyCheckInRecord, float64, error)
	Claim(ctx context.Context, claim DailyCheckInClaim) (*DailyCheckInRecord, bool, error)
	ListForAdmin(ctx context.Context, filter DailyCheckInAdminFilter) ([]DailyCheckInAdminRecord, int64, error)
	ListForUser(ctx context.Context, userID int64, page, pageSize int) ([]DailyCheckInRecord, int64, error)
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
	reward             func() (float64, error)
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

func secureDailyCheckInReward() (float64, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(dailyCheckInRewardRollCount))
	if err != nil {
		return 0, fmt.Errorf("generate daily check-in reward: %w", err)
	}
	return dailyCheckInRewardFromRoll(value.Int64())
}

func dailyCheckInRewardFromRoll(roll int64) (float64, error) {
	if roll < 0 || roll >= dailyCheckInRewardRollCount {
		return 0, fmt.Errorf("daily check-in reward roll must be between 0 and 99")
	}
	if roll < 50 {
		return 0.25, nil
	}
	if roll < 80 {
		return 0.5, nil
	}
	return 1, nil
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
	if reward != 0.25 && reward != 0.5 && reward != 1 {
		return nil, fmt.Errorf("daily check-in reward must be one of 0.25, 0.5, or 1")
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

// ListForUserBalanceHistory exposes the immutable daily ledger in the common
// balance-history shape. It is read-only and deliberately excludes recharge
// accounting fields such as total_recharged.
func (s *DailyCheckInService) ListForUserBalanceHistory(
	ctx context.Context,
	userID int64,
	page, pageSize int,
) ([]RedeemCode, int64, error) {
	if userID <= 0 {
		return nil, 0, ErrUserNotFound
	}
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("daily check-in repository is unavailable")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	records, total, err := s.repo.ListForUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list daily check-in balance history: %w", err)
	}
	items := make([]RedeemCode, 0, len(records))
	for _, record := range records {
		usedBy := record.UserID
		usedAt := record.CheckedInAt
		items = append(items, RedeemCode{
			ID:        record.ID,
			Code:      fmt.Sprintf("CHECKIN-%d", record.ID),
			Type:      RedeemTypeDailyCheckIn,
			Value:     record.RewardAmount,
			Status:    StatusUsed,
			UsedBy:    &usedBy,
			UsedAt:    &usedAt,
			CreatedAt: record.CreatedAt,
		})
	}
	return items, total, nil
}

// ListForAdmin returns immutable ledger rows without participating in claims.
func (s *DailyCheckInService) ListForAdmin(ctx context.Context, filter DailyCheckInAdminFilter) (*DailyCheckInAdminList, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("daily check-in repository is unavailable")
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	} else if filter.PageSize > 200 {
		filter.PageSize = 200
	}
	filter.Query = strings.TrimSpace(filter.Query)
	filter.ServiceDate = strings.TrimSpace(filter.ServiceDate)

	if filter.AllDates && filter.ServiceDate != "" {
		return nil, infraerrors.BadRequest(
			"DAILY_CHECK_IN_DATE_MODE_CONFLICT",
			"service_date cannot be combined with all dates",
		)
	}
	if filter.ServiceDate != "" {
		if parsed, err := time.Parse(time.DateOnly, filter.ServiceDate); err != nil || parsed.Format(time.DateOnly) != filter.ServiceDate {
			return nil, infraerrors.BadRequest("DAILY_CHECK_IN_SERVICE_DATE_INVALID", "service_date must use YYYY-MM-DD")
		}
	} else if !filter.AllDates {
		filter.ServiceDate, _ = dailyCheckInWindow(s.now())
	}

	items, total, err := s.repo.ListForAdmin(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list daily check-ins for admin: %w", err)
	}
	if items == nil {
		items = []DailyCheckInAdminRecord{}
	}
	return &DailyCheckInAdminList{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
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
