-- Add per-user privacy filter configuration.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS privacy_filter_config TEXT NOT NULL DEFAULT '{"enabled":false,"types":["ip_address","email","phone","id_card","bank_card","api_key","token","private_key","random_string"]}';
