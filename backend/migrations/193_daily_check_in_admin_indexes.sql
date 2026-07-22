-- Support the administrator's default service-day view and all-history pagination.

CREATE INDEX IF NOT EXISTS idx_daily_check_ins_service_date_checked_at_id
    ON daily_check_ins (service_date, checked_in_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_daily_check_ins_checked_at_id
    ON daily_check_ins (checked_in_at DESC, id DESC);
