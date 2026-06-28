package service

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestGwRefundRequiresManualReferenceForPersonalQRCode(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	order := &dbent.PaymentOrder{
		ID:             601,
		Amount:         18,
		PayAmount:      18,
		PaymentType:    payment.TypeAlipay,
		ProviderKey:    psNilIfEmpty(payment.TypePersonalQR),
		OutTradeNo:     "sub2_personal_refund_order",
		PaymentTradeNo: "sub2_personal_refund_order",
	}
	svc := &PaymentService{
		entClient: client,
	}

	err := svc.gwRefund(ctx, &RefundPlan{
		OrderID:       order.ID,
		Order:         order,
		RefundAmount:  order.Amount,
		GatewayAmount: order.PayAmount,
		Reason:        "customer refund",
	})
	require.Error(t, err)
	require.Equal(t, "MANUAL_REFUND_REFERENCE_REQUIRED", infraerrors.Reason(err))
}

func TestGwRefundAuditsManualReferenceForPersonalQRCode(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	order := &dbent.PaymentOrder{
		ID:             602,
		Amount:         18,
		PayAmount:      18,
		PaymentType:    payment.TypeAlipay,
		ProviderKey:    psNilIfEmpty(payment.TypePersonalQR),
		OutTradeNo:     "sub2_personal_refund_order_confirmed",
		PaymentTradeNo: "sub2_personal_refund_order_confirmed",
	}
	svc := &PaymentService{
		entClient: client,
	}

	err := svc.gwRefund(ctx, &RefundPlan{
		OrderID:               order.ID,
		Order:                 order,
		RefundAmount:          order.Amount,
		GatewayAmount:         order.PayAmount,
		Reason:                "customer refund",
		ManualRefundReference: "alipay-refund-20260626",
	})
	require.NoError(t, err)

	logs, err := svc.GetOrderAuditLogs(ctx, order.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	require.Equal(t, "REFUND_MANUAL_ORIGINAL_ROUTE_CONFIRMED", logs[0].Action)
	require.Contains(t, logs[0].Detail, "alipay-refund-20260626")
}
