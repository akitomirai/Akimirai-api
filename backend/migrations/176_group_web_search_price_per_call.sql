-- Codex alpha/search web search per-call price override.
-- NULL uses the built-in default 0.01 USD/call; 0 means free.
ALTER TABLE groups ADD COLUMN IF NOT EXISTS web_search_price_per_call DECIMAL(20,8);
