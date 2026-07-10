ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS request_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS request_total_ms INTEGER,
    ADD COLUMN IF NOT EXISTS request_body_read_ms INTEGER,
    ADD COLUMN IF NOT EXISTS request_body_bytes BIGINT,
    ADD COLUMN IF NOT EXISTS upstream_request_written_ms INTEGER,
    ADD COLUMN IF NOT EXISTS upstream_first_byte_ms INTEGER,
    ADD COLUMN IF NOT EXISTS request_first_token_ms INTEGER,
    ADD COLUMN IF NOT EXISTS route_kind VARCHAR(16),
    ADD COLUMN IF NOT EXISTS proxy_id_snapshot BIGINT,
    ADD COLUMN IF NOT EXISTS proxy_name_snapshot VARCHAR(255),
    ADD COLUMN IF NOT EXISTS proxy_protocol_snapshot VARCHAR(16),
    ADD COLUMN IF NOT EXISTS route_fingerprint VARCHAR(64),
    ADD COLUMN IF NOT EXISTS final_upstream_status INTEGER,
    ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS account_switch_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS attempt_timeline JSONB;

CREATE INDEX IF NOT EXISTS idx_usage_logs_proxy_snapshot_created_at
    ON usage_logs (proxy_id_snapshot, created_at DESC)
    WHERE proxy_id_snapshot IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_usage_logs_route_kind_created_at
    ON usage_logs (route_kind, created_at DESC)
    WHERE route_kind IS NOT NULL;
