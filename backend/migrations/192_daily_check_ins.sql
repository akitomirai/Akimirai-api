-- Durable exact-once ledger for the daily balance reward.

CREATE TABLE IF NOT EXISTS daily_check_ins (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_date    DATE NOT NULL,
    reward_amount   DECIMAL(20, 8) NOT NULL CHECK (reward_amount IN (1, 2, 3)),
    balance_before  DECIMAL(20, 8) NOT NULL,
    balance_after   DECIMAL(20, 8) NOT NULL,
    checked_in_at   TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_daily_check_ins_user_service_date UNIQUE (user_id, service_date),
    CONSTRAINT chk_daily_check_ins_balance_snapshot
        CHECK (balance_after = balance_before + reward_amount)
);

COMMENT ON TABLE daily_check_ins IS
    'Immutable exact-once daily check-in reward ledger; service days reset at 02:00 Asia/Shanghai';
