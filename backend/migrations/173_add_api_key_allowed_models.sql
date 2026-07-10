ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS allowed_models JSONB NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN api_keys.allowed_models IS 'Exact model allowlist for this API key; an empty array allows all models available to the assigned group';
