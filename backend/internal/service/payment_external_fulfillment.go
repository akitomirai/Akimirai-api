package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/externalfulfillmentsku"
	"github.com/Wei-Shaw/sub2api/ent/externalorderfulfillment"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	ExternalFulfillmentStatusPending      = "pending"
	ExternalFulfillmentStatusFulfilled    = "fulfilled"
	ExternalFulfillmentStatusNotifyFailed = "notify_failed"
	ExternalFulfillmentStatusFailed       = "failed"

	ExternalFulfillmentNotifySkipped = "skipped"
	ExternalFulfillmentNotifySent    = "sent"
	ExternalFulfillmentNotifyFailed  = "failed"

	externalFulfillmentDefaultPlatform = "xianyu"
	externalFulfillmentWebhookEnv      = "EXTERNAL_FULFILLMENT_FEISHU_WEBHOOK_URL"
)

var sendExternalFulfillmentFeishu = sendFeishuTextWebhook

type ExternalFulfillmentSKURequest struct {
	Platform         string  `json:"platform"`
	SKUCode          string  `json:"sku_code"`
	Name             string  `json:"name"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	RedeemType       string  `json:"redeem_type"`
	RedeemValue      float64 `json:"redeem_value"`
	GroupID          *int64  `json:"group_id"`
	ValidityDays     int     `json:"validity_days"`
	ExpiresInDays    *int    `json:"expires_in_days"`
	ManualURL        string  `json:"manual_url"`
	DeliveryTemplate string  `json:"delivery_template"`
	Enabled          *bool   `json:"enabled"`
}

type CreateExternalFulfillmentRequest struct {
	Platform        string  `json:"platform"`
	PlatformOrderID string  `json:"platform_order_id"`
	BuyerRef        string  `json:"buyer_ref"`
	SKUCode         string  `json:"sku_code"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	ManualURL       string  `json:"manual_url"`
	NotifyFeishu    *bool   `json:"notify_feishu"`
	Operator        string  `json:"operator"`
}

type ExternalFulfillmentListParams struct {
	Page      int
	PageSize  int
	Platform  string
	Status    string
	SKUCode   string
	Keyword   string
	Notify    string
	CreatedAt time.Time
}

type ExternalFulfillmentSKUListParams struct {
	Page     int
	PageSize int
	Platform string
	Enabled  *bool
	Keyword  string
}

type ExternalFulfillmentResult struct {
	Fulfillment *dbent.ExternalOrderFulfillment `json:"fulfillment"`
	Replay      bool                            `json:"replay"`
}

func (s *PaymentService) ListExternalFulfillmentSKUs(ctx context.Context, p ExternalFulfillmentSKUListParams) ([]*dbent.ExternalFulfillmentSKU, int, error) {
	if s == nil || s.entClient == nil {
		return nil, 0, infraerrors.InternalServer("PAYMENT_SERVICE_UNCONFIGURED", "payment service not configured")
	}
	q := s.entClient.ExternalFulfillmentSKU.Query()
	if platform := normalizeExternalPlatform(p.Platform); platform != "" {
		q = q.Where(externalfulfillmentsku.PlatformEQ(platform))
	}
	if p.Enabled != nil {
		q = q.Where(externalfulfillmentsku.EnabledEQ(*p.Enabled))
	}
	if keyword := strings.TrimSpace(p.Keyword); keyword != "" {
		q = q.Where(externalfulfillmentsku.Or(
			externalfulfillmentsku.SkuCodeContainsFold(keyword),
			externalfulfillmentsku.NameContainsFold(keyword),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count external fulfillment skus: %w", err)
	}
	ps, pg := applyPagination(p.PageSize, p.Page)
	items, err := q.Order(dbent.Desc(externalfulfillmentsku.FieldUpdatedAt)).Limit(ps).Offset((pg - 1) * ps).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("query external fulfillment skus: %w", err)
	}
	return items, total, nil
}

func (s *PaymentService) UpsertExternalFulfillmentSKU(ctx context.Context, req ExternalFulfillmentSKURequest) (*dbent.ExternalFulfillmentSKU, error) {
	if s == nil || s.entClient == nil {
		return nil, infraerrors.InternalServer("PAYMENT_SERVICE_UNCONFIGURED", "payment service not configured")
	}
	normalized, err := normalizeExternalFulfillmentSKURequest(req)
	if err != nil {
		return nil, err
	}
	existing, err := s.entClient.ExternalFulfillmentSKU.Query().
		Where(
			externalfulfillmentsku.PlatformEQ(normalized.Platform),
			externalfulfillmentsku.SkuCodeEQ(normalized.SKUCode),
		).
		Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return nil, fmt.Errorf("lookup external fulfillment sku: %w", err)
	}
	if existing != nil {
		up := s.entClient.ExternalFulfillmentSKU.UpdateOneID(existing.ID).
			SetName(normalized.Name).
			SetAmount(normalized.Amount).
			SetCurrency(normalized.Currency).
			SetRedeemType(normalized.RedeemType).
			SetRedeemValue(normalized.RedeemValue).
			SetValidityDays(normalized.ValidityDays).
			SetEnabled(normalized.Enabled != nil && *normalized.Enabled)
		setExternalSKUUpdateOptionals(up, normalized)
		item, err := up.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("update external fulfillment sku: %w", err)
		}
		return item, nil
	}
	create := s.entClient.ExternalFulfillmentSKU.Create().
		SetPlatform(normalized.Platform).
		SetSkuCode(normalized.SKUCode).
		SetName(normalized.Name).
		SetAmount(normalized.Amount).
		SetCurrency(normalized.Currency).
		SetRedeemType(normalized.RedeemType).
		SetRedeemValue(normalized.RedeemValue).
		SetValidityDays(normalized.ValidityDays).
		SetEnabled(normalized.Enabled != nil && *normalized.Enabled)
	create.SetNillableGroupID(normalized.GroupID)
	create.SetNillableExpiresInDays(normalized.ExpiresInDays)
	create.SetNillableManualURL(optionalStringPtr(normalized.ManualURL))
	create.SetNillableDeliveryTemplate(optionalStringPtr(normalized.DeliveryTemplate))
	item, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create external fulfillment sku: %w", err)
	}
	return item, nil
}

func (s *PaymentService) DeleteExternalFulfillmentSKU(ctx context.Context, id int64) error {
	if s == nil || s.entClient == nil {
		return infraerrors.InternalServer("PAYMENT_SERVICE_UNCONFIGURED", "payment service not configured")
	}
	if id <= 0 {
		return infraerrors.BadRequest("EXTERNAL_FULFILLMENT_SKU_ID_INVALID", "sku id must be positive")
	}
	if err := s.entClient.ExternalFulfillmentSKU.DeleteOneID(id).Exec(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return infraerrors.NotFound("EXTERNAL_FULFILLMENT_SKU_NOT_FOUND", "sku not found")
		}
		return fmt.Errorf("delete external fulfillment sku: %w", err)
	}
	return nil
}

func (s *PaymentService) CreateExternalFulfillment(ctx context.Context, req CreateExternalFulfillmentRequest) (*ExternalFulfillmentResult, error) {
	if s == nil || s.entClient == nil {
		return nil, infraerrors.InternalServer("PAYMENT_SERVICE_UNCONFIGURED", "payment service not configured")
	}
	req.Platform = normalizeExternalPlatform(req.Platform)
	req.PlatformOrderID = strings.TrimSpace(req.PlatformOrderID)
	req.SKUCode = strings.TrimSpace(req.SKUCode)
	req.BuyerRef = strings.TrimSpace(req.BuyerRef)
	req.ManualURL = strings.TrimSpace(req.ManualURL)
	req.Currency = normalizeExternalCurrency(req.Currency)
	req.Operator = strings.TrimSpace(req.Operator)

	if req.PlatformOrderID == "" {
		return nil, infraerrors.BadRequest("EXTERNAL_ORDER_ID_REQUIRED", "platform_order_id is required")
	}
	if req.SKUCode == "" {
		return nil, infraerrors.BadRequest("EXTERNAL_SKU_CODE_REQUIRED", "sku_code is required")
	}

	existing, err := s.entClient.ExternalOrderFulfillment.Query().
		Where(
			externalorderfulfillment.PlatformEQ(req.Platform),
			externalorderfulfillment.PlatformOrderID(req.PlatformOrderID),
		).
		Only(ctx)
	if err == nil {
		return &ExternalFulfillmentResult{Fulfillment: existing, Replay: true}, nil
	}
	if !dbent.IsNotFound(err) {
		return nil, fmt.Errorf("lookup external fulfillment: %w", err)
	}

	sku, err := s.getEnabledExternalFulfillmentSKU(ctx, req.Platform, req.SKUCode)
	if err != nil {
		return nil, err
	}
	amount := req.Amount
	if amount == 0 {
		amount = sku.Amount
	}
	currency := req.Currency
	if currency == "" {
		currency = normalizeExternalCurrency(sku.Currency)
	}
	manualURL := req.ManualURL
	if manualURL == "" && sku.ManualURL != nil {
		manualURL = strings.TrimSpace(*sku.ManualURL)
	}
	expiresAt := externalFulfillmentExpiresAt(sku.ExpiresInDays)
	deliveryMessage := buildExternalDeliveryMessage(sku.DeliveryTemplate, sku, req, "", manualURL)

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start external fulfillment tx: %w", err)
	}
	txCtx := dbent.NewTxContext(ctx, tx)
	defer func() { _ = tx.Rollback() }()

	record, err := tx.ExternalOrderFulfillment.Create().
		SetPlatform(req.Platform).
		SetPlatformOrderID(req.PlatformOrderID).
		SetSkuCode(req.SKUCode).
		SetNillableBuyerRef(optionalStringPtr(req.BuyerRef)).
		SetNillableSkuName(optionalStringPtr(sku.Name)).
		SetAmount(amount).
		SetCurrency(currency).
		SetRedeemType(sku.RedeemType).
		SetRedeemValue(sku.RedeemValue).
		SetNillableGroupID(sku.GroupID).
		SetValidityDays(sku.ValidityDays).
		SetNillableExpiresAt(expiresAt).
		SetNillableManualURL(optionalStringPtr(manualURL)).
		SetNillableDeliveryMessage(optionalStringPtr(deliveryMessage)).
		SetStatus(ExternalFulfillmentStatusPending).
		SetNotifyStatus(ExternalFulfillmentNotifySkipped).
		SetNillableOperator(optionalStringPtr(req.Operator)).
		Save(txCtx)
	if err != nil {
		return nil, fmt.Errorf("create external fulfillment record: %w", err)
	}

	code, err := s.createExternalFulfillmentRedeemCode(txCtx, record, sku, expiresAt, deliveryMessage)
	if err != nil {
		return nil, err
	}
	deliveryMessage = buildExternalDeliveryMessage(sku.DeliveryTemplate, sku, req, code.Code, manualURL)
	now := time.Now()
	record, err = tx.ExternalOrderFulfillment.UpdateOneID(record.ID).
		SetRedeemCodeID(code.ID).
		SetRedeemCode(code.Code).
		SetDeliveryMessage(deliveryMessage).
		SetStatus(ExternalFulfillmentStatusFulfilled).
		SetDeliveredAt(now).
		SetNotifyStatus(ExternalFulfillmentNotifySkipped).
		ClearFailReason().
		Save(txCtx)
	if err != nil {
		return nil, fmt.Errorf("update external fulfillment record: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit external fulfillment tx: %w", err)
	}

	notify := req.NotifyFeishu == nil || *req.NotifyFeishu
	if notify {
		record = s.notifyExternalFulfillmentBestEffort(ctx, record)
	}
	return &ExternalFulfillmentResult{Fulfillment: record}, nil
}

func (s *PaymentService) ListExternalFulfillments(ctx context.Context, p ExternalFulfillmentListParams) ([]*dbent.ExternalOrderFulfillment, int, error) {
	if s == nil || s.entClient == nil {
		return nil, 0, infraerrors.InternalServer("PAYMENT_SERVICE_UNCONFIGURED", "payment service not configured")
	}
	q := s.entClient.ExternalOrderFulfillment.Query()
	if platform := normalizeExternalPlatform(p.Platform); platform != "" {
		q = q.Where(externalorderfulfillment.PlatformEQ(platform))
	}
	if status := strings.TrimSpace(p.Status); status != "" {
		q = q.Where(externalorderfulfillment.StatusEQ(status))
	}
	if skuCode := strings.TrimSpace(p.SKUCode); skuCode != "" {
		q = q.Where(externalorderfulfillment.SkuCodeEQ(skuCode))
	}
	if notify := strings.TrimSpace(p.Notify); notify != "" {
		q = q.Where(externalorderfulfillment.NotifyStatusEQ(notify))
	}
	if keyword := strings.TrimSpace(p.Keyword); keyword != "" {
		q = q.Where(externalorderfulfillment.Or(
			externalorderfulfillment.PlatformOrderIDContainsFold(keyword),
			externalorderfulfillment.BuyerRefContainsFold(keyword),
			externalorderfulfillment.RedeemCodeContainsFold(keyword),
			externalorderfulfillment.SkuCodeContainsFold(keyword),
		))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count external fulfillments: %w", err)
	}
	ps, pg := applyPagination(p.PageSize, p.Page)
	items, err := q.Order(dbent.Desc(externalorderfulfillment.FieldCreatedAt)).Limit(ps).Offset((pg - 1) * ps).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("query external fulfillments: %w", err)
	}
	return items, total, nil
}

func (s *PaymentService) RetryExternalFulfillmentNotify(ctx context.Context, id int64) (*dbent.ExternalOrderFulfillment, error) {
	if s == nil || s.entClient == nil {
		return nil, infraerrors.InternalServer("PAYMENT_SERVICE_UNCONFIGURED", "payment service not configured")
	}
	if id <= 0 {
		return nil, infraerrors.BadRequest("EXTERNAL_FULFILLMENT_ID_INVALID", "fulfillment id must be positive")
	}
	record, err := s.entClient.ExternalOrderFulfillment.Get(ctx, id)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.NotFound("EXTERNAL_FULFILLMENT_NOT_FOUND", "fulfillment not found")
		}
		return nil, err
	}
	if strings.TrimSpace(externalFulfillmentDeliveryMessage(record)) == "" {
		return nil, infraerrors.BadRequest("EXTERNAL_FULFILLMENT_MESSAGE_EMPTY", "delivery message is empty")
	}
	return s.notifyExternalFulfillment(ctx, record)
}

func (s *PaymentService) getEnabledExternalFulfillmentSKU(ctx context.Context, platform, skuCode string) (*dbent.ExternalFulfillmentSKU, error) {
	sku, err := s.entClient.ExternalFulfillmentSKU.Query().
		Where(
			externalfulfillmentsku.PlatformEQ(platform),
			externalfulfillmentsku.SkuCodeEQ(skuCode),
			externalfulfillmentsku.EnabledEQ(true),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.NotFound("EXTERNAL_FULFILLMENT_SKU_NOT_FOUND", "enabled sku not found")
		}
		return nil, fmt.Errorf("lookup external fulfillment sku: %w", err)
	}
	return sku, nil
}

func (s *PaymentService) createExternalFulfillmentRedeemCode(ctx context.Context, record *dbent.ExternalOrderFulfillment, sku *dbent.ExternalFulfillmentSKU, expiresAt *time.Time, deliveryMessage string) (*RedeemCode, error) {
	if s.redeemService == nil {
		return nil, infraerrors.InternalServer("EXTERNAL_FULFILLMENT_REDEEM_UNCONFIGURED", "redeem service not configured")
	}
	tx := dbent.TxFromContext(ctx)
	if tx == nil {
		return nil, infraerrors.InternalServer("EXTERNAL_FULFILLMENT_TX_REQUIRED", "external fulfillment transaction missing")
	}
	codeValue, err := generateExternalFulfillmentRedeemCode()
	if err != nil {
		return nil, fmt.Errorf("generate external fulfillment redeem code: %w", err)
	}
	notes := fmt.Sprintf("external fulfillment: platform=%s order=%s sku=%s", record.Platform, record.PlatformOrderID, record.SkuCode)
	if deliveryMessage != "" {
		notes += "\n" + deliveryMessage
	}
	code := &RedeemCode{
		Code:         codeValue,
		Type:         sku.RedeemType,
		Value:        sku.RedeemValue,
		Status:       StatusUnused,
		Notes:        notes,
		GroupID:      sku.GroupID,
		ValidityDays: sku.ValidityDays,
		ExpiresAt:    expiresAt,
	}
	created, err := tx.RedeemCode.Create().
		SetCode(code.Code).
		SetType(code.Type).
		SetValue(code.Value).
		SetStatus(code.Status).
		SetNotes(code.Notes).
		SetValidityDays(code.ValidityDays).
		SetNillableGroupID(code.GroupID).
		SetNillableExpiresAt(code.ExpiresAt).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create external fulfillment redeem code: %w", err)
	}
	code.ID = created.ID
	code.CreatedAt = created.CreatedAt
	return code, nil
}

func (s *PaymentService) notifyExternalFulfillmentBestEffort(ctx context.Context, record *dbent.ExternalOrderFulfillment) *dbent.ExternalOrderFulfillment {
	updated, err := s.notifyExternalFulfillment(ctx, record)
	if err != nil {
		return record
	}
	return updated
}

func (s *PaymentService) notifyExternalFulfillment(ctx context.Context, record *dbent.ExternalOrderFulfillment) (*dbent.ExternalOrderFulfillment, error) {
	webhookURL := strings.TrimSpace(os.Getenv(externalFulfillmentWebhookEnv))
	if webhookURL == "" {
		return s.entClient.ExternalOrderFulfillment.UpdateOneID(record.ID).
			SetNotifyStatus(ExternalFulfillmentNotifySkipped).
			ClearFailReason().
			Save(ctx)
	}
	message := strings.TrimSpace(externalFulfillmentDeliveryMessage(record))
	if message == "" {
		message = "Redeem code: " + strings.TrimSpace(externalFulfillmentRedeemCode(record))
	}
	if err := sendExternalFulfillmentFeishu(ctx, webhookURL, message); err != nil {
		reason := err.Error()
		return s.entClient.ExternalOrderFulfillment.UpdateOneID(record.ID).
			SetStatus(ExternalFulfillmentStatusNotifyFailed).
			SetNotifyStatus(ExternalFulfillmentNotifyFailed).
			SetFailReason(reason).
			Save(ctx)
	}
	now := time.Now()
	return s.entClient.ExternalOrderFulfillment.UpdateOneID(record.ID).
		SetStatus(ExternalFulfillmentStatusFulfilled).
		SetNotifyStatus(ExternalFulfillmentNotifySent).
		SetNotifiedAt(now).
		ClearFailReason().
		Save(ctx)
}

func normalizeExternalFulfillmentSKURequest(req ExternalFulfillmentSKURequest) (ExternalFulfillmentSKURequest, error) {
	req.Platform = normalizeExternalPlatform(req.Platform)
	req.SKUCode = strings.TrimSpace(req.SKUCode)
	req.Name = strings.TrimSpace(req.Name)
	req.Currency = normalizeExternalCurrency(req.Currency)
	req.RedeemType = strings.TrimSpace(req.RedeemType)
	req.ManualURL = strings.TrimSpace(req.ManualURL)
	req.DeliveryTemplate = strings.TrimSpace(req.DeliveryTemplate)
	if req.SKUCode == "" {
		return req, infraerrors.BadRequest("EXTERNAL_SKU_CODE_REQUIRED", "sku_code is required")
	}
	if req.Name == "" {
		return req, infraerrors.BadRequest("EXTERNAL_SKU_NAME_REQUIRED", "name is required")
	}
	if err := validateExternalFulfillmentRedeem(req.RedeemType, req.RedeemValue, req.GroupID, req.ValidityDays); err != nil {
		return req, err
	}
	if req.Enabled == nil {
		enabled := true
		req.Enabled = &enabled
	}
	return req, nil
}

func validateExternalFulfillmentRedeem(redeemType string, value float64, groupID *int64, validityDays int) error {
	switch redeemType {
	case domain.RedeemTypeBalance, domain.RedeemTypeConcurrency:
		if value == 0 {
			return infraerrors.BadRequest("EXTERNAL_REDEEM_VALUE_REQUIRED", "redeem_value must not be zero")
		}
	case domain.RedeemTypeSubscription:
		if groupID == nil || *groupID <= 0 {
			return infraerrors.BadRequest("EXTERNAL_REDEEM_GROUP_REQUIRED", "group_id is required for subscription redeem")
		}
		if validityDays == 0 {
			return infraerrors.BadRequest("EXTERNAL_REDEEM_VALIDITY_REQUIRED", "validity_days must not be zero for subscription redeem")
		}
	case domain.RedeemTypeInvitation:
	default:
		return infraerrors.BadRequest("EXTERNAL_REDEEM_TYPE_INVALID", "redeem_type is invalid")
	}
	return nil
}

func setExternalSKUUpdateOptionals(up *dbent.ExternalFulfillmentSKUUpdateOne, req ExternalFulfillmentSKURequest) {
	if req.GroupID != nil {
		up.SetGroupID(*req.GroupID)
	} else {
		up.ClearGroupID()
	}
	if req.ExpiresInDays != nil {
		up.SetExpiresInDays(*req.ExpiresInDays)
	} else {
		up.ClearExpiresInDays()
	}
	if manualURL := optionalStringPtr(req.ManualURL); manualURL != nil {
		up.SetManualURL(*manualURL)
	} else {
		up.ClearManualURL()
	}
	if tmpl := optionalStringPtr(req.DeliveryTemplate); tmpl != nil {
		up.SetDeliveryTemplate(*tmpl)
	} else {
		up.ClearDeliveryTemplate()
	}
}

func externalFulfillmentExpiresAt(expiresInDays *int) *time.Time {
	if expiresInDays == nil || *expiresInDays <= 0 {
		return nil
	}
	t := time.Now().UTC().AddDate(0, 0, *expiresInDays)
	return &t
}

func generateExternalFulfillmentRedeemCode() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

func buildExternalDeliveryMessage(template *string, sku *dbent.ExternalFulfillmentSKU, req CreateExternalFulfillmentRequest, code, manualURL string) string {
	tmpl := ""
	if template != nil {
		tmpl = strings.TrimSpace(*template)
	}
	if tmpl == "" {
		tmpl = "Your Bakaai redeem code: {{code}}\nOrder: {{order_id}}\nPlan: {{sku_name}}\nManual: {{manual_url}}"
	}
	replacements := map[string]string{
		"{{code}}":       code,
		"{{order_id}}":   req.PlatformOrderID,
		"{{platform}}":   req.Platform,
		"{{buyer_ref}}":  req.BuyerRef,
		"{{sku_code}}":   req.SKUCode,
		"{{sku_name}}":   sku.Name,
		"{{manual_url}}": manualURL,
		"{{amount}}":     strconv.FormatFloat(firstNonZeroFloat(req.Amount, sku.Amount), 'f', -1, 64),
		"{{currency}}":   firstNonEmptyExternalString(req.Currency, sku.Currency),
	}
	out := tmpl
	for old, newValue := range replacements {
		out = strings.ReplaceAll(out, old, newValue)
	}
	return strings.TrimSpace(out)
}

func sendFeishuTextWebhook(ctx context.Context, webhookURL, text string) error {
	payload := map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": text,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("feishu webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func normalizeExternalPlatform(platform string) string {
	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform == "" {
		return externalFulfillmentDefaultPlatform
	}
	return platform
}

func normalizeExternalCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return "CNY"
	}
	return currency
}

func optionalStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func firstNonZeroFloat(values ...float64) float64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstNonEmptyExternalString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func externalFulfillmentDeliveryMessage(e *dbent.ExternalOrderFulfillment) string {
	if e == nil || e.DeliveryMessage == nil {
		return ""
	}
	return *e.DeliveryMessage
}

func externalFulfillmentRedeemCode(e *dbent.ExternalOrderFulfillment) string {
	if e == nil || e.RedeemCode == nil {
		return ""
	}
	return *e.RedeemCode
}
