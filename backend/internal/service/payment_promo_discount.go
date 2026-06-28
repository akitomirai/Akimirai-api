package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/promocode"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

const paymentPromoAppliedAuditAction = "PROMO_CODE_PAYMENT_APPLIED"

type paymentPromoDiscount struct {
	Code    string
	Percent float64
}

type calculatedPaymentPromoDiscount struct {
	Code           string
	Percent        float64
	BaseAmount     float64
	DiscountAmount float64
}

type ValidatePaymentPromoCodeRequest struct {
	Code        string
	Amount      float64
	PaymentType string
	OrderType   string
	PlanID      int64
}

type PaymentPromoCodeQuote struct {
	Code             string  `json:"code"`
	DiscountPercent  float64 `json:"discount_percent"`
	DiscountAmount   float64 `json:"discount_amount"`
	DiscountedAmount float64 `json:"discounted_amount"`
}

func (s *PaymentService) ValidatePaymentPromoCode(ctx context.Context, req ValidatePaymentPromoCodeRequest) (*PaymentPromoCodeQuote, error) {
	if req.OrderType == "" {
		req.OrderType = payment.OrderTypeBalance
	}
	if normalized := NormalizeVisibleMethod(req.PaymentType); normalized != "" {
		req.PaymentType = normalized
	}
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get payment config: %w", err)
	}
	if !cfg.Enabled {
		return nil, infraerrors.Forbidden("PAYMENT_DISABLED", "payment system is disabled")
	}

	plan, err := s.validateOrderInput(ctx, CreateOrderRequest{
		Amount:      req.Amount,
		PaymentType: req.PaymentType,
		OrderType:   req.OrderType,
		PlanID:      req.PlanID,
	}, cfg)
	if err != nil {
		return nil, err
	}

	limitAmount := req.Amount
	if plan != nil {
		limitAmount = plan.Price
	}

	currency := payment.DefaultPaymentCurrency
	if s.configService != nil && strings.TrimSpace(req.PaymentType) != "" {
		currency, err = s.configService.ValidateMethodCurrencyConsistency(ctx, req.PaymentType)
		if err != nil {
			return nil, err
		}
	}

	discount, err := s.resolvePaymentPromoDiscount(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	applied, err := calculatePaymentPromoDiscount(limitAmount, discount, currency)
	if err != nil {
		return nil, err
	}
	if applied == nil {
		return nil, ErrPromoCodeNoDiscount
	}
	return &PaymentPromoCodeQuote{
		Code:             applied.Code,
		DiscountPercent:  applied.Percent,
		DiscountAmount:   applied.DiscountAmount,
		DiscountedAmount: applied.BaseAmount,
	}, nil
}

func (s *PaymentService) resolvePaymentPromoDiscount(ctx context.Context, rawCode string) (*paymentPromoDiscount, error) {
	code := strings.ToUpper(strings.TrimSpace(rawCode))
	if code == "" {
		return nil, nil
	}

	promo, err := s.entClient.PromoCode.Query().
		Where(promocode.CodeEqualFold(code)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrPromoCodeNotFound
		}
		return nil, err
	}
	if promo.Status != PromoCodeStatusActive {
		return nil, ErrPromoCodeDisabled
	}
	if promo.ExpiresAt != nil && time.Now().After(*promo.ExpiresAt) {
		return nil, ErrPromoCodeExpired
	}
	if promo.MaxUses > 0 && promo.UsedCount >= promo.MaxUses {
		return nil, ErrPromoCodeMaxUsed
	}
	if promo.DiscountPercent <= 0 {
		return nil, ErrPromoCodeNoDiscount
	}
	if promo.DiscountPercent >= 100 || math.IsNaN(promo.DiscountPercent) || math.IsInf(promo.DiscountPercent, 0) {
		return nil, ErrPromoCodeInvalid
	}
	return &paymentPromoDiscount{Code: strings.ToUpper(promo.Code), Percent: promo.DiscountPercent}, nil
}

func calculatePaymentPromoDiscount(baseAmount float64, discount *paymentPromoDiscount, currency string) (*calculatedPaymentPromoDiscount, error) {
	if discount == nil || discount.Percent <= 0 {
		return nil, nil
	}
	if err := validateCreateOrderAmountCurrency(baseAmount, currency); err != nil {
		return nil, err
	}

	fractionDigits := int32(payment.CurrencyMaxFractionDigits(currency))
	original := decimal.NewFromFloat(baseAmount)
	multiplier := decimal.NewFromFloat(100 - discount.Percent).Div(decimal.NewFromInt(100))
	discounted := original.Mul(multiplier).Round(fractionDigits)
	if !discounted.GreaterThan(decimal.Zero) {
		return nil, ErrPromoCodeInvalid
	}
	discountAmount := original.Sub(discounted).Round(fractionDigits)
	baseAmountStr := discounted.StringFixed(fractionDigits)
	if _, err := payment.AmountToMinorUnit(baseAmountStr, currency); err != nil {
		return nil, ErrPromoCodeInvalid
	}
	base, err := strconv.ParseFloat(baseAmountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("parse discounted payment amount: %w", err)
	}
	discountValue, err := strconv.ParseFloat(discountAmount.StringFixed(fractionDigits), 64)
	if err != nil {
		return nil, fmt.Errorf("parse payment discount amount: %w", err)
	}
	return &calculatedPaymentPromoDiscount{
		Code:           discount.Code,
		Percent:        discount.Percent,
		BaseAmount:     base,
		DiscountAmount: math.Max(0, discountValue),
	}, nil
}

func (s *PaymentService) applyPaymentPromoUsage(ctx context.Context, order *dbent.PaymentOrder) error {
	if s == nil || order == nil || strings.TrimSpace(psStringValue(order.PromoCode)) == "" || order.DiscountPercent <= 0 {
		return nil
	}
	if s.hasAuditLog(ctx, order.ID, paymentPromoAppliedAuditAction) {
		return nil
	}
	promo, err := s.entClient.PromoCode.Query().
		Where(promocode.CodeEqualFold(psStringValue(order.PromoCode))).
		Only(ctx)
	if err != nil {
		s.writeAuditLog(ctx, order.ID, "PROMO_CODE_PAYMENT_USAGE_FAILED", "system", map[string]any{
			"promoCode": psStringValue(order.PromoCode),
			"error":     err.Error(),
		})
		return nil
	}
	if _, err := s.entClient.PromoCode.UpdateOneID(promo.ID).AddUsedCount(1).Save(ctx); err != nil {
		s.writeAuditLog(ctx, order.ID, "PROMO_CODE_PAYMENT_USAGE_FAILED", "system", map[string]any{
			"promoCode": promo.Code,
			"error":     err.Error(),
		})
		return nil
	}
	s.writeAuditLog(ctx, order.ID, paymentPromoAppliedAuditAction, "system", map[string]any{
		"promoCode":       promo.Code,
		"discountPercent": order.DiscountPercent,
		"discountAmount":  order.DiscountAmount,
	})
	return nil
}
