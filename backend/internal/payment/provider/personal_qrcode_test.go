package provider

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestPersonalQRCodeCreatePaymentReturnsConfiguredQRCode(t *testing.T) {
	t.Parallel()

	prov, err := NewPersonalQRCode("1", map[string]string{
		"alipayQr": "https://example.com/alipay-qr",
		"wxpayQr":  "https://example.com/wxpay-qr",
	})
	if err != nil {
		t.Fatalf("NewPersonalQRCode returned error: %v", err)
	}

	resp, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "sub2_test",
		PaymentType: payment.TypeWxpay,
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if resp.TradeNo != "sub2_test" {
		t.Fatalf("TradeNo = %q, want %q", resp.TradeNo, "sub2_test")
	}
	if resp.QRCode != "https://example.com/wxpay-qr" {
		t.Fatalf("QRCode = %q", resp.QRCode)
	}
}

func TestPersonalQRCodeRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	prov, err := NewPersonalQRCode("1", map[string]string{"alipayQr": "alipay"})
	if err != nil {
		t.Fatalf("NewPersonalQRCode returned error: %v", err)
	}

	if _, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "sub2_test",
		PaymentType: payment.TypeStripe,
	}); err == nil {
		t.Fatal("expected unsupported payment type error")
	}
}

func TestPersonalQRCodeRefundRequiresManualFlow(t *testing.T) {
	t.Parallel()

	prov, err := NewPersonalQRCode("1", map[string]string{"alipayQr": "alipay"})
	if err != nil {
		t.Fatalf("NewPersonalQRCode returned error: %v", err)
	}

	if _, err := prov.Refund(context.Background(), payment.RefundRequest{}); err == nil {
		t.Fatal("expected manual refund error")
	}
}
