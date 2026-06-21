-- Add latency observability fields captured by the OpenAI/Responses gateway.
-- The fields are nullable so existing usage records remain valid.
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS client_transport VARCHAR(10);
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS auth_latency_ms INTEGER;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS routing_latency_ms INTEGER;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS upstream_latency_ms INTEGER;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS response_latency_ms INTEGER;
