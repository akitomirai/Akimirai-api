-- Keep historical integer rewards readable while allowing the new decimal schedule.

ALTER TABLE daily_check_ins
    DROP CONSTRAINT IF EXISTS daily_check_ins_reward_amount_check,
    DROP CONSTRAINT IF EXISTS chk_daily_check_ins_reward_amount;

ALTER TABLE daily_check_ins
    ADD CONSTRAINT chk_daily_check_ins_reward_amount
        CHECK (reward_amount IN (0.25, 0.5, 1, 2, 3));
