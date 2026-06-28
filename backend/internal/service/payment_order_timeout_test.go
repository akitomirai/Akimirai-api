package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestPaymentOrderTimeoutMinutes(t *testing.T) {
	t.Parallel()

	require.Equal(t, 5, paymentOrderTimeoutMinutes(
		&PaymentConfig{OrderTimeoutMin: 30},
		&payment.InstanceSelection{ProviderKey: payment.TypePersonalQR},
	))
	require.Equal(t, 12, paymentOrderTimeoutMinutes(
		&PaymentConfig{OrderTimeoutMin: 12},
		&payment.InstanceSelection{ProviderKey: payment.TypeAlipay},
	))
	require.Equal(t, defaultOrderTimeoutMin, paymentOrderTimeoutMinutes(&PaymentConfig{}, nil))
}
