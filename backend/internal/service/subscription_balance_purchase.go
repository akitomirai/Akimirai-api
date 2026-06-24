package service

import (
	"context"
	"fmt"
	"math"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type BalanceSubscriptionPurchaseInput struct {
	UserID int64
	PlanID int64
}

type BalanceSubscriptionPurchaseResult struct {
	Subscription  *UserSubscription
	BalanceBefore float64
	BalanceAfter  float64
	Price         float64
	Shortfall     float64
}

// PurchaseSubscriptionWithBalance deducts user balance and assigns the selected
// subscription plan in one local transaction. It intentionally does not create
// payment orders or call payment providers.
func (s *PaymentService) PurchaseSubscriptionWithBalance(ctx context.Context, input BalanceSubscriptionPurchaseInput) (*BalanceSubscriptionPurchaseResult, error) {
	if s == nil || s.entClient == nil || s.subscriptionSvc == nil || s.configService == nil {
		return nil, infraerrors.ServiceUnavailable("BALANCE_SUBSCRIPTION_UNAVAILABLE", "balance subscription purchase is unavailable")
	}
	if input.UserID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_USER", "user is required")
	}
	if input.PlanID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "subscription plan is required")
	}

	plan, err := s.validateSubOrder(ctx, CreateOrderRequest{
		UserID:    input.UserID,
		OrderType: payment.OrderTypeSubscription,
		PlanID:    input.PlanID,
	})
	if err != nil {
		return nil, err
	}
	price := normalizeSubscriptionBalanceAmount(plan.Price)
	if price <= 0 || math.IsNaN(price) || math.IsInf(price, 0) {
		return nil, infraerrors.BadRequest("PLAN_PRICE_INVALID", "plan price must be positive")
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin balance subscription transaction: %w", err)
	}
	txCtx := dbent.NewTxContext(ctx, tx)
	defer func() { _ = tx.Rollback() }()

	userEntity, err := tx.User.Get(txCtx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if userEntity.Status != payment.EntityStatusActive {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "user account is disabled")
	}

	balanceBefore := normalizeSubscriptionBalanceAmount(userEntity.Balance)
	if balanceBefore+0.0000001 < price {
		return nil, insufficientSubscriptionBalanceError(balanceBefore, price)
	}

	affected, err := tx.User.Update().
		Where(
			dbuser.IDEQ(input.UserID),
			dbuser.StatusEQ(payment.EntityStatusActive),
			dbuser.BalanceGTE(price),
		).
		AddBalance(-price).
		Save(txCtx)
	if err != nil {
		return nil, fmt.Errorf("deduct balance: %w", err)
	}
	if affected == 0 {
		currentBalance := balanceBefore
		if current, getErr := tx.User.Get(txCtx, input.UserID); getErr == nil {
			currentBalance = normalizeSubscriptionBalanceAmount(current.Balance)
		}
		return nil, insufficientSubscriptionBalanceError(currentBalance, price)
	}

	sub, _, err := s.subscriptionSvc.AssignOrExtendSubscription(txCtx, &AssignSubscriptionInput{
		UserID:       input.UserID,
		GroupID:      plan.GroupID,
		ValidityDays: psComputeValidityDays(plan.ValidityDays, plan.ValidityUnit),
		Notes:        fmt.Sprintf("balance subscription plan %d", plan.ID),
	})
	if err != nil {
		return nil, err
	}

	updatedUser, err := tx.User.Get(txCtx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get updated user: %w", err)
	}
	balanceAfter := normalizeSubscriptionBalanceAmount(updatedUser.Balance)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit balance subscription transaction: %w", err)
	}

	return &BalanceSubscriptionPurchaseResult{
		Subscription:  sub,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Price:         price,
		Shortfall:     0,
	}, nil
}

func insufficientSubscriptionBalanceError(balance, price float64) error {
	shortfall := normalizeSubscriptionBalanceAmount(price - balance)
	if shortfall < 0 {
		shortfall = 0
	}
	return infraerrors.BadRequest("INSUFFICIENT_BALANCE", "insufficient balance").
		WithMetadata(map[string]string{
			"balance":   fmt.Sprintf("%.2f", normalizeSubscriptionBalanceAmount(balance)),
			"price":     fmt.Sprintf("%.2f", normalizeSubscriptionBalanceAmount(price)),
			"shortfall": fmt.Sprintf("%.2f", shortfall),
		})
}

func normalizeSubscriptionBalanceAmount(value float64) float64 {
	return math.Round(value*100) / 100
}
