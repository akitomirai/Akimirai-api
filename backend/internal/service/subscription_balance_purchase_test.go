package service

import (
	"context"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPurchaseSubscriptionWithBalanceDeductsBalanceAndAssignsSubscription(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("balance-sub@example.com").
		SetPasswordHash("hash").
		SetStatus(StatusActive).
		SetBalance(100).
		Save(ctx)
	require.NoError(t, err)

	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(10).
		SetName("Monthly Pro").
		SetPrice(25).
		SetValidityDays(1).
		SetValidityUnit("month").
		SetForSale(true).
		Save(ctx)
	require.NoError(t, err)

	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 10, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	subscriptionSvc := NewSubscriptionService(groupRepo, subRepo, nil, client, nil)
	paymentSvc := &PaymentService{
		entClient:       client,
		configService:   &PaymentConfigService{entClient: client},
		subscriptionSvc: subscriptionSvc,
		groupRepo:       groupRepo,
	}

	result, err := paymentSvc.PurchaseSubscriptionWithBalance(ctx, BalanceSubscriptionPurchaseInput{
		UserID: user.ID,
		PlanID: plan.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Subscription)
	require.Equal(t, int64(10), result.Subscription.GroupID)
	require.Equal(t, 100.0, result.BalanceBefore)
	require.Equal(t, 75.0, result.BalanceAfter)
	require.Equal(t, 25.0, result.Price)
	require.Equal(t, 1, subRepo.createCalls)

	updatedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 75.0, updatedUser.Balance)
}

func TestPurchaseSubscriptionWithBalanceRejectsInsufficientBalance(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("low-balance-sub@example.com").
		SetPasswordHash("hash").
		SetStatus(StatusActive).
		SetBalance(10).
		Save(ctx)
	require.NoError(t, err)

	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(10).
		SetName("Monthly Pro").
		SetPrice(25).
		SetValidityDays(30).
		SetValidityUnit("day").
		SetForSale(true).
		Save(ctx)
	require.NoError(t, err)

	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 10, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	paymentSvc := &PaymentService{
		entClient:       client,
		configService:   &PaymentConfigService{entClient: client},
		subscriptionSvc: NewSubscriptionService(groupRepo, subRepo, nil, client, nil),
		groupRepo:       groupRepo,
	}

	result, err := paymentSvc.PurchaseSubscriptionWithBalance(ctx, BalanceSubscriptionPurchaseInput{
		UserID: user.ID,
		PlanID: plan.ID,
	})
	require.Nil(t, result)
	require.Error(t, err)
	appErr := infraerrors.FromError(err)
	require.Equal(t, "INSUFFICIENT_BALANCE", appErr.Reason)
	require.Equal(t, "15.00", appErr.Metadata["shortfall"])
	require.Equal(t, 0, subRepo.createCalls)

	updatedUser, err := client.User.Get(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 10.0, updatedUser.Balance)
}
