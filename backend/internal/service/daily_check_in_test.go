package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type dailyCheckInRepositoryStub struct {
	mu             sync.Mutex
	record         *DailyCheckInRecord
	currentBalance float64
	created        bool
	err            error
	getDate        string
	claim          DailyCheckInClaim
	claimCalls     int
}

func (s *dailyCheckInRepositoryStub) GetForServiceDate(_ context.Context, _ int64, serviceDate string) (*DailyCheckInRecord, float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getDate = serviceDate
	return s.record, s.currentBalance, s.err
}

func (s *dailyCheckInRepositoryStub) Claim(_ context.Context, claim DailyCheckInClaim) (*DailyCheckInRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.claim = claim
	s.claimCalls++
	return s.record, s.created, s.err
}

type dailyCheckInAuthInvalidatorStub struct {
	mu      sync.Mutex
	userIDs []int64
}

func (*dailyCheckInAuthInvalidatorStub) InvalidateAuthCacheByKey(context.Context, string)    {}
func (*dailyCheckInAuthInvalidatorStub) InvalidateAuthCacheByGroupID(context.Context, int64) {}
func (s *dailyCheckInAuthInvalidatorStub) InvalidateAuthCacheByUserID(_ context.Context, userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userIDs = append(s.userIDs, userID)
}

type dailyCheckInBalanceInvalidatorStub struct {
	mu      sync.Mutex
	userIDs []int64
	err     error
}

func (s *dailyCheckInBalanceInvalidatorStub) InvalidateUserBalance(_ context.Context, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userIDs = append(s.userIDs, userID)
	return s.err
}

func mustParseCheckInTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
	return parsed
}

func TestDailyCheckInServiceUsesTwoAMShanghaiBoundary(t *testing.T) {
	tests := []struct {
		name          string
		now           string
		serviceDate   string
		nextResetTime string
	}{
		{
			name:          "one second before reset belongs to previous service day",
			now:           "2026-07-21T01:59:59+08:00",
			serviceDate:   "2026-07-20",
			nextResetTime: "2026-07-21T02:00:00+08:00",
		},
		{
			name:          "exact reset starts the new service day",
			now:           "2026-07-21T02:00:00+08:00",
			serviceDate:   "2026-07-21",
			nextResetTime: "2026-07-22T02:00:00+08:00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &dailyCheckInRepositoryStub{currentBalance: 9.5}
			svc := newDailyCheckInService(repo, nil, nil)
			svc.now = func() time.Time { return mustParseCheckInTime(t, tc.now) }

			status, err := svc.GetStatus(context.Background(), 17)
			require.NoError(t, err)
			require.False(t, status.CheckedIn)
			require.False(t, status.AlreadyCheckedIn)
			require.Equal(t, tc.serviceDate, status.ServiceDate)
			require.Equal(t, tc.serviceDate, repo.getDate)
			require.Equal(t, tc.nextResetTime, status.NextResetAt.Format(time.RFC3339))
			require.Equal(t, 9.5, status.BalanceBefore)
			require.Equal(t, 9.5, status.BalanceAfter)
			require.Nil(t, status.CheckedInAt)
		})
	}
}

func TestDailyCheckInServiceClaimReturnsPersistedRewardAndInvalidatesCaches(t *testing.T) {
	checkedAt := mustParseCheckInTime(t, "2026-07-21T03:04:05+08:00")
	record := &DailyCheckInRecord{
		ID:            41,
		UserID:        7,
		ServiceDate:   "2026-07-21",
		RewardAmount:  2,
		BalanceBefore: 10,
		BalanceAfter:  12,
		CheckedInAt:   checkedAt,
	}
	repo := &dailyCheckInRepositoryStub{record: record, created: true}
	authInvalidator := &dailyCheckInAuthInvalidatorStub{}
	balanceInvalidator := &dailyCheckInBalanceInvalidatorStub{err: errors.New("redis unavailable")}
	svc := newDailyCheckInService(repo, authInvalidator, balanceInvalidator)
	svc.now = func() time.Time { return checkedAt }
	svc.reward = func() (int, error) { return 2, nil }

	status, err := svc.Claim(context.Background(), 7)
	require.NoError(t, err, "cache invalidation is best effort")
	require.True(t, status.CheckedIn)
	require.False(t, status.AlreadyCheckedIn)
	require.Equal(t, float64(2), status.RewardAmount)
	require.Equal(t, float64(10), status.BalanceBefore)
	require.Equal(t, float64(12), status.BalanceAfter)
	require.Equal(t, checkedAt, *status.CheckedInAt)
	require.Equal(t, float64(2), repo.claim.RewardAmount)
	require.Equal(t, "2026-07-21", repo.claim.ServiceDate)
	require.Equal(t, checkedAt, repo.claim.CheckedInAt)
	require.Equal(t, []int64{7}, authInvalidator.userIDs)
	require.Equal(t, []int64{7}, balanceInvalidator.userIDs)
}

func TestDailyCheckInServiceReplayDoesNotInvalidateCachesAgain(t *testing.T) {
	checkedAt := mustParseCheckInTime(t, "2026-07-21T03:04:05+08:00")
	repo := &dailyCheckInRepositoryStub{
		record: &DailyCheckInRecord{
			UserID: 7, ServiceDate: "2026-07-21", RewardAmount: 1,
			BalanceBefore: 10, BalanceAfter: 11, CheckedInAt: checkedAt,
		},
		created: false,
	}
	authInvalidator := &dailyCheckInAuthInvalidatorStub{}
	balanceInvalidator := &dailyCheckInBalanceInvalidatorStub{}
	svc := newDailyCheckInService(repo, authInvalidator, balanceInvalidator)
	svc.now = func() time.Time { return checkedAt }
	svc.reward = func() (int, error) { return 3, nil }

	status, err := svc.Claim(context.Background(), 7)
	require.NoError(t, err)
	require.True(t, status.CheckedIn)
	require.True(t, status.AlreadyCheckedIn)
	require.Equal(t, float64(1), status.RewardAmount, "replay must return the original persisted reward")
	require.Empty(t, authInvalidator.userIDs)
	require.Empty(t, balanceInvalidator.userIDs)
}

func TestDailyCheckInServiceRejectsInvalidRewardSource(t *testing.T) {
	repo := &dailyCheckInRepositoryStub{}
	svc := newDailyCheckInService(repo, nil, nil)
	svc.reward = func() (int, error) { return 4, nil }

	_, err := svc.Claim(context.Background(), 7)
	require.ErrorContains(t, err, "reward")
	require.Zero(t, repo.claimCalls)
}
