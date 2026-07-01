CREATE TABLE IF NOT EXISTS external_fulfillment_skus (
    id BIGSERIAL PRIMARY KEY,
    platform VARCHAR(32) NOT NULL DEFAULT 'xianyu',
    sku_code VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    redeem_type VARCHAR(32) NOT NULL,
    redeem_value DECIMAL(20,6) NOT NULL DEFAULT 0,
    group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    validity_days INTEGER NOT NULL DEFAULT 0,
    expires_in_days INTEGER,
    manual_url TEXT,
    delivery_template TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT external_fulfillment_skus_platform_code_unique UNIQUE (platform, sku_code)
);

CREATE TABLE IF NOT EXISTS external_order_fulfillments (
    id BIGSERIAL PRIMARY KEY,
    platform VARCHAR(32) NOT NULL DEFAULT 'xianyu',
    platform_order_id VARCHAR(128) NOT NULL,
    buyer_ref VARCHAR(255),
    sku_code VARCHAR(128) NOT NULL,
    sku_name VARCHAR(255),
    amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
    redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    redeem_code VARCHAR(128),
    redeem_type VARCHAR(32) NOT NULL,
    redeem_value DECIMAL(20,6) NOT NULL DEFAULT 0,
    group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    validity_days INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    manual_url TEXT,
    delivery_message TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    notify_status VARCHAR(32) NOT NULL DEFAULT 'skipped',
    fail_reason TEXT,
    operator VARCHAR(128),
    delivered_at TIMESTAMPTZ,
    notified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT external_order_fulfillments_platform_order_unique UNIQUE (platform, platform_order_id)
);

CREATE INDEX IF NOT EXISTS idx_external_fulfillment_skus_platform_enabled
    ON external_fulfillment_skus (platform, enabled);

CREATE INDEX IF NOT EXISTS idx_external_order_fulfillments_status
    ON external_order_fulfillments (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_external_order_fulfillments_platform_sku
    ON external_order_fulfillments (platform, sku_code, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_external_order_fulfillments_redeem_code_id
    ON external_order_fulfillments (redeem_code_id);

COMMENT ON TABLE external_fulfillment_skus IS 'Marketplace SKU to redeem-code fulfillment mapping.';
COMMENT ON TABLE external_order_fulfillments IS 'External marketplace order fulfillment records and delivery state.';
