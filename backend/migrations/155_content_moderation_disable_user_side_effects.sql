-- Disable content moderation user-facing side effects for existing installs.
-- Risk-control should block/record by default; email notices and auto-ban must
-- be explicit admin opt-ins.
UPDATE settings
SET value = (
        COALESCE(NULLIF(value, ''), '{}')::jsonb
        || jsonb_build_object(
            'email_on_hit', false,
            'auto_ban_enabled', false
        )
    )::text,
    updated_at = NOW()
WHERE key = 'content_moderation_config';
