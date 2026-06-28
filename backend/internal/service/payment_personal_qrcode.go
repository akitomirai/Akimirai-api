package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type ConfirmPersonalQRCodePaymentRequest struct {
	Amount    float64
	Method    string
	ReceiptID string
	Note      string
	Operator  string
}

const personalQRCodeOrderTimeoutMin = 5

func (s *PaymentService) ConfirmPersonalQRCodePayment(ctx context.Context, orderID int64, req ConfirmPersonalQRCodePaymentRequest) (*dbent.PaymentOrder, error) {
	order, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	operator := strings.TrimSpace(req.Operator)
	if operator == "" {
		operator = "admin"
	}
	method := NormalizeVisibleMethod(req.Method)
	if method != payment.TypeAlipay && method != payment.TypeWxpay {
		return nil, infraerrors.BadRequest("INVALID_PAYMENT_METHOD", "payment method must be alipay or wxpay")
	}
	if NormalizeVisibleMethod(order.PaymentType) != method {
		return nil, infraerrors.BadRequest("PAYMENT_METHOD_MISMATCH", "payment method does not match order")
	}
	if !paymentOrderUsesProvider(order, payment.TypePersonalQR) {
		return nil, infraerrors.BadRequest("INVALID_PROVIDER", "order is not a personal QR-code payment")
	}
	if strings.TrimSpace(req.ReceiptID) == "" {
		return nil, infraerrors.BadRequest("RECEIPT_REQUIRED", "receipt reference is required")
	}
	if !isValidProviderAmount(req.Amount) {
		return nil, infraerrors.BadRequest("INVALID_AMOUNT", "invalid paid amount")
	}
	if math.Abs(req.Amount-order.PayAmount) > paymentAmountToleranceForCurrency(PaymentOrderCurrency(order)) {
		return nil, infraerrors.BadRequest("PAYMENT_AMOUNT_MISMATCH", "paid amount does not match order")
	}
	if !personalQRCodeOrderAllowsManualConfirm(order) {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "order cannot be manually confirmed in current status")
	}

	detail := map[string]any{
		"method":    method,
		"amount":    req.Amount,
		"receiptId": strings.TrimSpace(req.ReceiptID),
		"note":      strings.TrimSpace(req.Note),
		"source":    "admin_manual_personal_qrcode",
	}
	s.writeAuditLog(ctx, order.ID, "PAYMENT_MANUAL_CONFIRM_ATTEMPT", operator, detail)

	tradeNo := strings.TrimSpace(req.ReceiptID)
	if tradeNo == "" {
		tradeNo = order.OutTradeNo
	}
	if err := s.confirmPayment(ctx, order.ID, tradeNo, req.Amount, payment.TypePersonalQR, nil); err != nil {
		return nil, err
	}
	s.writeAuditLog(ctx, order.ID, "PAYMENT_MANUAL_CONFIRMED", operator, detail)

	updated, err := s.entClient.PaymentOrder.Get(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("reload order: %w", err)
	}
	return updated, nil
}

func personalQRCodeOrderAllowsManualConfirm(order *dbent.PaymentOrder) bool {
	if order == nil {
		return false
	}
	switch order.Status {
	case OrderStatusPending, OrderStatusCancelled:
		return true
	case OrderStatusExpired:
		return order.UpdatedAt.After(time.Now().Add(-paymentGraceMinutes * time.Minute))
	default:
		return false
	}
}

func paymentOrderUsesProvider(order *dbent.PaymentOrder, providerKey string) bool {
	if order == nil {
		return false
	}
	providerKey = strings.TrimSpace(providerKey)
	if providerKey == "" {
		return false
	}
	if snapshot := psOrderProviderSnapshot(order); snapshot != nil && strings.EqualFold(strings.TrimSpace(snapshot.ProviderKey), providerKey) {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(psStringValue(order.ProviderKey)), providerKey)
}
