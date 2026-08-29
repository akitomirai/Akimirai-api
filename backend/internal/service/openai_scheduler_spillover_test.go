package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayService_LoadAwarenessStickySpilloverDoesNotReturnFullStickyWaitPlan(t *testing.T) {
	groupID := int64(8801)
	sessionHash := "sticky-spillover-load-error"
	accounts := []Account{
		{ID: 88011, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 0, GroupIDs: []int64{groupID}},
		{ID: 88012, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 1, GroupIDs: []int64{groupID}},
	}
	cache := &schedulerTestGatewayCache{sessionBindings: map[string]int64{"openai:" + sessionHash: accounts[0].ID}}
	concurrency := schedulerTestConcurrencyCache{
		loadBatchErr:   errors.New("load batch unavailable"),
		acquireResults: map[int64]bool{accounts[0].ID: false, accounts[1].ID: false},
		waitCounts:     map[int64]int{accounts[0].ID: 1},
	}
	cfg := &config.Config{RunMode: config.RunModeStandard}
	cfg.Gateway.Scheduling.LoadBatchEnabled = true
	cfg.Gateway.Scheduling.StickySessionMaxWaiting = 1

	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: accounts},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(concurrency),
	}

	selection, err := svc.SelectAccountWithLoadAwareness(context.Background(), &groupID, sessionHash, "gpt-5.1", nil)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.WaitPlan)
	require.Equal(t, accounts[1].ID, selection.WaitPlan.AccountID)
	require.NotEqual(t, accounts[0].ID, selection.WaitPlan.AccountID,
		"fallback waiting must not enqueue on the sticky account whose queue is full")
}
