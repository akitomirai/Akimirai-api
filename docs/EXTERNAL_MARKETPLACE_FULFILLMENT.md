# External Marketplace Fulfillment

This module supports marketplace resale workflows such as Xianyu order intake:

1. Configure a SKU mapping from marketplace SKU to Bakaai redeem-code package.
2. Submit the paid marketplace order.
3. The backend creates one redeem code and stores a fulfillment record.
4. The backend optionally sends the delivery message to Feishu.
5. Admins can list fulfillment state and retry Feishu notification.

## Feishu Webhook

Set the webhook URL in the backend runtime environment:

```bash
EXTERNAL_FULFILLMENT_FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/...
```

If the variable is not set, fulfillment still succeeds and `notify_status` is `skipped`.

## Admin APIs

All endpoints are under the existing admin payment group and require admin auth.

### Upsert SKU

`POST /api/v1/admin/payment/external-fulfillment-skus`

```json
{
  "platform": "xianyu",
  "sku_code": "bakaai-30",
  "name": "Bakaai 30 balance",
  "amount": 30,
  "currency": "CNY",
  "redeem_type": "balance",
  "redeem_value": 30,
  "manual_url": "https://example.com/bakaai-manual",
  "delivery_template": "您的 Bakaai 兑换码：{{code}}\n订单号：{{order_id}}\n操作手册：{{manual_url}}",
  "enabled": true
}
```

For subscription codes, pass `redeem_type: "subscription"`, `group_id`, and `validity_days`.

### List SKUs

`GET /api/v1/admin/payment/external-fulfillment-skus?platform=xianyu&enabled=true&page=1&page_size=20`

### Submit Order

`POST /api/v1/admin/payment/external-fulfillments`

```json
{
  "platform": "xianyu",
  "platform_order_id": "XY202607010001",
  "buyer_ref": "buyer nickname or masked id",
  "sku_code": "bakaai-30",
  "notify_feishu": true
}
```

The same `platform + platform_order_id` is idempotent. Repeating the request returns the existing fulfillment and does not create another redeem code.

### List Fulfillments

`GET /api/v1/admin/payment/external-fulfillments?platform=xianyu&status=fulfilled&keyword=XY202607`

Useful statuses:

- `fulfilled`: code created and delivery message ready.
- `notify_failed`: code created, but Feishu notification failed.
- `failed`: fulfillment failed before delivery.

Useful notify states:

- `sent`: Feishu webhook accepted the message.
- `skipped`: webhook URL was not configured or notification was disabled.
- `failed`: webhook call failed.

### Retry Feishu Notification

`POST /api/v1/admin/payment/external-fulfillments/:id/retry-notify`

This only retries the message delivery. It does not create a new redeem code.
