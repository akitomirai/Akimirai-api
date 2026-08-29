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
	adminFilter    DailyCheckInAdminFilter
	adminItems     []DailyCheckInAdminRecord
	adminTotal     int64
	userItems      []DailyCheckInRecord
	userTotal      int64
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

func (s *dailyCheckInRepositoryStub) ListForAdmin(_ context.Context, filter DailyCheckInAdminFilter) ([]DailyCheckInAdminRecord, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adminFilter = filter
	return s.adminItems, s.adminTotal, s.err
}

func (s *dailyCheckInRepositoryStub) ListForUser(_ context.Context, _ int64, _, _ int) ([]DailyCheckInRecord, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.userItems, s.userTotal, s.err
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

func TestDailyCheckInRewardUsesWeightedDistribution(t *testing.T) {
	tests := []struct {
		name   string
		roll   int64
		reward float64
	}{
		{name: "quarter bucket starts at zero", roll: 0, reward: 0.25},
		{name: "quarter bucket ends at forty nine", roll: 49, reward: 0.25},
		{name: "half bucket starts at fifty", roll: 50, reward: 0.5},
		{name: "half bucket ends at seventy nine", roll: 79, reward: 0.5},
		{name: "one bucket starts at eighty", roll: 80, reward: 1},
		{name: "last bucket awards one", roll: 99, reward: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reward, err := dailyCheckInRewardFromRoll(tc.roll)
			require.NoError(t, err)
			require.Equal(t, tc.reward, reward)
		})
	}

	counts := map[float64]int{}
	for roll := int64(0); roll < 100; roll++ {
		reward, err := dailyCheckInRewardFromRoll(roll)
		require.NoError(t, err)
		counts[reward]++
	}
	require.Equal(t, map[float64]int{0.25: 50, 0.5: 30, 1: 20}, counts)

	for _, invalidRoll := range []int64{-1, 100} {
		_, err := dailyCheckInRewardFromRoll(invalidRoll)
		require.ErrorContains(t, err, "roll")
	}
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
		RewardAmount:  0.5,
		BalanceBefore: 10,
		BalanceAfter:  12,
		CheckedInAt:   checkedAt,
	}
	repo := &dailyCheckInRepositoryStub{record: record, created: true}
	authInvalidator := &dailyCheckInAuthInvalidatorStub{}
	balanceInvalidator := &dailyCheckInBalanceInvalidatorStub{err: errors.New("redis unavailable")}
	svc := newDailyCheckInService(repo, authInvalidator, balanceInvalidator)
	svc.now = func() time.Time { return checkedAt }
	svc.reward = func() (float64, error) { return 0.5, nil }

	status, err := svc.Claim(context.Background(), 7)
	require.NoError(t, err, "cache invalidation is best effort")
	require.True(t, status.CheckedIn)
	require.False(t, status.AlreadyCheckedIn)
	require.Equal(t, float64(0.5), status.RewardAmount)
	require.Equal(t, float64(10), status.BalanceBefore)
	require.Equal(t, float64(12), status.BalanceAfter)
	require.Equal(t, checkedAt, *status.CheckedInAt)
	require.Equal(t, float64(0.5), repo.claim.RewardAmount)
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
	svc.reward = func() (float64, error) { return 1, nil }

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
	svc.reward = func() (float64, error) { return 2, nil }

	_, err := svc.Claim(context.Background(), 7)
	require.ErrorContains(t, err, "reward")
	require.Zero(t, repo.claimCalls)
}

func TestDailyCheckInServiceProjectsBalanceHistory(t *testing.T) {
	checkedAt := mustParseCheckInTime(t, "2026-07-21T03:04:05+08:00")
	repo := &dailyCheckInRepositoryStub{
		userItems: []DailyCheckInRecord{{
			ID: 41, UserID: 7, RewardAmount: 0.25, CheckedInAt: checkedAt, CreatedAt: checkedAt,
		}},
		userTotal: 1,
	}
	svc := newDailyCheckInService(repo, nil, nil)

	items, total, err := svc.ListForUserBalanceHistory(context.Background(), 7, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, int64(41), items[0].ID)
	require.Equal(t, "CHECKIN-41", items[0].Code)
	require.Equal(t, RedeemTypeDailyCheckIn, items[0].Type)
	require.Equal(t, 0.25, items[0].Value)
}

func TestDailyCheckInServiceAdminListDefaultsToCurrentServiceDay(t *testing.T) {
	repo := &dailyCheckInRepositoryStub{
		adminItems: []DailyCheckInAdminRecord{{ID: 41, UserID: 7, ServiceDate: "2026-07-20"}},
		adminTotal: 1,
	}
	svc := newDailyCheckInService(repo, nil, nil)
	svc.now = func() time.Time { return mustParseCheckInTime(t, "2026-07-21T01:59:59+08:00") }

	result, err := svc.ListForAdmin(context.Background(), DailyCheckInAdminFilter{
		Page: 0, PageSize: 500, Query: "  user@example.com  ",
	})
	require.NoError(t, err)
	require.Equal(t, "2026-07-20", repo.adminFilter.ServiceDate)
	require.False(t, repo.adminFilter.AllDates)
	require.Equal(t, 1, repo.adminFilter.Page)
	require.Equal(t, 200, repo.adminFilter.PageSize)
	require.Equal(t, "user@example.com", repo.adminFilter.Query)
	require.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
}

func TestDailyCheckInServiceAdminListAllDatesDoesNotApplyDefaultDate(t *testing.T) {
	repo := &dailyCheckInRepositoryStub{}
	svc := newDailyCheckInService(repo, nil, nil)
	svc.now = func() time.Time { return mustParseCheckInTime(t, "2026-07-21T03:00:00+08:00") }

	_, err := svc.ListForAdmin(context.Background(), DailyCheckInAdminFilter{AllDates: true})
	require.NoError(t, err)
	require.True(t, repo.adminFilter.AllDates)
	require.Empty(t, repo.adminFilter.ServiceDate)
}

func TestDailyCheckInServiceAdminListRejectsConflictingDateMode(t *testing.T) {
	repo := &dailyCheckInRepositoryStub{}
	svc := newDailyCheckInService(repo, nil, nil)

	_, err := svc.ListForAdmin(context.Background(), DailyCheckInAdminFilter{
		AllDates: true, ServiceDate: "2026-07-21",
	})
	require.ErrorContains(t, err, "all dates")
	require.Empty(t, repo.adminFilter.ServiceDate)
}
