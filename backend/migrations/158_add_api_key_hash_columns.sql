-- Add non-destructive API key hash columns for one-time-display keys.
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS key_hash VARCHAR(96),
    ADD COLUMN IF NOT EXISTS key_prefix VARCHAR(32);

ALTER TABLE deleted_api_key_audits
    ADD COLUMN IF NOT EXISTS key_hash VARCHAR(96),
    ADD COLUMN IF NOT EXISTS key_prefix VARCHAR(32);

CREATE UNIQUE INDEX IF NOT EXISTS api_keys_key_hash_unique
    ON api_keys (key_hash)
    WHERE key_hash IS NOT NULL AND key_hash <> '';

CREATE INDEX IF NOT EXISTS api_keys_key_prefix_idx
    ON api_keys (key_prefix)
    WHERE key_prefix IS NOT NULL AND key_prefix <> '';

CREATE INDEX IF NOT EXISTS deletedapikeyaudit_key_hash
    ON deleted_api_key_audits (key_hash)
    WHERE key_hash IS NOT NULL AND key_hash <> '';
