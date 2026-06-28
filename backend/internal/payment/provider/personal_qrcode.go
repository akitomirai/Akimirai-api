package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	personalQRCodeAlipayQR = "alipayQr"
	personalQRCodeWxpayQR  = "wxpayQr"
)

// PersonalQRCode serves static personal collection codes. It intentionally does
// not verify payment or refund upstream; admins must confirm receipts manually.
type PersonalQRCode struct {
	instanceID string
	config     map[string]string
}

func NewPersonalQRCode(instanceID string, config map[string]string) (*PersonalQRCode, error) {
	if strings.TrimSpace(config[personalQRCodeAlipayQR]) == "" && strings.TrimSpace(config[personalQRCodeWxpayQR]) == "" {
		return nil, fmt.Errorf("personal_qrcode config missing required key: alipayQr or wxpayQr")
	}
	cfg := make(map[string]string, len(config))
	for k, v := range config {
		cfg[k] = v
	}
	return &PersonalQRCode{instanceID: instanceID, config: cfg}, nil
}

func (p *PersonalQRCode) Name() string        { return "Personal QR Code" }
func (p *PersonalQRCode) ProviderKey() string { return payment.TypePersonalQR }
func (p *PersonalQRCode) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeAlipay, payment.TypeWxpay}
}

func (p *PersonalQRCode) MerchantIdentityMetadata() map[string]string {
	if p == nil {
		return nil
	}
	out := map[string]string{}
	if account := strings.TrimSpace(p.config["accountName"]); account != "" {
		out["account_name"] = account
	}
	if p.instanceID != "" {
		out["provider_instance_id"] = p.instanceID
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (p *PersonalQRCode) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	_ = ctx
	qr, err := p.qrForPaymentType(req.PaymentType)
	if err != nil {
		return nil, err
	}
	return &payment.CreatePaymentResponse{
		TradeNo:    req.OrderID,
		QRCode:     qr,
		Currency:   payment.DefaultPaymentCurrency,
		ResultType: payment.CreatePaymentResultOrderCreated,
	}, nil
}

func (p *PersonalQRCode) qrForPaymentType(paymentType string) (string, error) {
	switch payment.GetBasePaymentType(strings.TrimSpace(paymentType)) {
	case payment.TypeAlipay:
		if qr := strings.TrimSpace(p.config[personalQRCodeAlipayQR]); qr != "" {
			return qr, nil
		}
		return "", fmt.Errorf("personal_qrcode alipayQr is not configured")
	case payment.TypeWxpay:
		if qr := strings.TrimSpace(p.config[personalQRCodeWxpayQR]); qr != "" {
			return qr, nil
		}
		return "", fmt.Errorf("personal_qrcode wxpayQr is not configured")
	default:
		return "", fmt.Errorf("personal_qrcode unsupported payment type: %s", paymentType)
	}
}

func (p *PersonalQRCode) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	_ = ctx
	return &payment.QueryOrderResponse{
		TradeNo: strings.TrimSpace(tradeNo),
		Status:  payment.ProviderStatusPending,
	}, nil
}

func (p *PersonalQRCode) VerifyNotification(ctx context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	_, _, _ = ctx, rawBody, headers
	return nil, nil
}

func (p *PersonalQRCode) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	_, _ = ctx, req
	return nil, fmt.Errorf("personal_qrcode refunds require manual original-route confirmation")
}
