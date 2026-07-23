ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS prompt_cache_key_hash VARCHAR(64),
    ADD COLUMN IF NOT EXISTS prompt_cache_key_source VARCHAR(32),
    ADD COLUMN IF NOT EXISTS prompt_cache_prefix_hash VARCHAR(64),
    ADD COLUMN IF NOT EXISTS prompt_cache_tools_hash VARCHAR(64),
    ADD COLUMN IF NOT EXISTS prompt_cache_system_hash VARCHAR(64);

COMMENT ON COLUMN usage_logs.prompt_cache_key_hash IS
    'SHA-256 of the effective prompt cache identity; raw keys are never stored';
COMMENT ON COLUMN usage_logs.prompt_cache_key_source IS
    'Cache identity origin: none, client_header, client_body, or compat_derived';
COMMENT ON COLUMN usage_logs.prompt_cache_prefix_hash IS
    'SHA-256 of the normalized reusable prompt prefix';
COMMENT ON COLUMN usage_logs.prompt_cache_tools_hash IS
    'SHA-256 of normalized tools/functions definitions';
COMMENT ON COLUMN usage_logs.prompt_cache_system_hash IS
    'SHA-256 of normalized instructions and ordered system/developer messages';
