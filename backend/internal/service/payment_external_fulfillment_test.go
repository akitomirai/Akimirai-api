package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func newExternalFulfillmentTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	dbName := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
	db, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestCreateExternalFulfillmentCreatesRedeemCodeAndIsIdempotent(t *testing.T) {
	t.Setenv(externalFulfillmentWebhookEnv, "")

	ctx := context.Background()
	client := newExternalFulfillmentTestClient(t)
	svc := &PaymentService{entClient: client, redeemService: &RedeemService{}}

	enabled := true
	sku, err := svc.UpsertExternalFulfillmentSKU(ctx, ExternalFulfillmentSKURequest{
		Platform:         "xianyu",
		SKUCode:          "bakaai-30",
		Name:             "Bakaai 30 balance",
		Amount:           30,
		Currency:         "cny",
		RedeemType:       domain.RedeemTypeBalance,
		RedeemValue:      30,
		DeliveryTemplate: "code={{code}} manual={{manual_url}} order={{order_id}}",
		ManualURL:        "https://example.com/manual",
		Enabled:          &enabled,
	})
	require.NoError(t, err)
	require.Equal(t, "bakaai-30", sku.SkuCode)

	result, err := svc.CreateExternalFulfillment(ctx, CreateExternalFulfillmentRequest{
		PlatformOrderID: "XY202607010001",
		SKUCode:         "bakaai-30",
		BuyerRef:        "buyer-a",
	})
	require.NoError(t, err)
	require.False(t, result.Replay)
	require.NotNil(t, result.Fulfillment.RedeemCode)
	require.Equal(t, ExternalFulfillmentStatusFulfilled, result.Fulfillment.Status)
	require.Equal(t, ExternalFulfillmentNotifySkipped, result.Fulfillment.NotifyStatus)
	require.Contains(t, *result.Fulfillment.DeliveryMessage, *result.Fulfillment.RedeemCode)
	require.Contains(t, *result.Fulfillment.DeliveryMessage, "https://example.com/manual")

	code, err := client.RedeemCode.Query().Where(redeemcode.CodeEQ(*result.Fulfillment.RedeemCode)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, domain.RedeemTypeBalance, code.Type)
	require.Equal(t, 30.0, code.Value)

	replay, err := svc.CreateExternalFulfillment(ctx, CreateExternalFulfillmentRequest{
		PlatformOrderID: "XY202607010001",
		SKUCode:         "bakaai-30",
	})
	require.NoError(t, err)
	require.True(t, replay.Replay)
	require.Equal(t, result.Fulfillment.ID, replay.Fulfillment.ID)

	totalCodes, err := client.RedeemCode.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, totalCodes)
}

func TestCreateExternalFulfillmentRejectsUnknownSKU(t *testing.T) {
	ctx := context.Background()
	client := newExternalFulfillmentTestClient(t)
	svc := &PaymentService{entClient: client, redeemService: &RedeemService{}}

	_, err := svc.CreateExternalFulfillment(ctx, CreateExternalFulfillmentRequest{
		PlatformOrderID: "XY-MISSING",
		SKUCode:         "missing",
	})
	require.Error(t, err)
	require.Equal(t, "EXTERNAL_FULFILLMENT_SKU_NOT_FOUND", infraerrors.Reason(err))
}
