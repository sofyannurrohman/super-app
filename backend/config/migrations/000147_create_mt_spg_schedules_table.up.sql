CREATE TABLE IF NOT EXISTS mt_spg_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    spg_id UUID NOT NULL REFERENCES mt_spgs(id) ON DELETE CASCADE,
    key_account_id UUID NOT NULL REFERENCES key_accounts(id) ON DELETE CASCADE,
    supermarket_name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    day_pattern VARCHAR(50) NOT NULL DEFAULT 'WEEKEND_FRI_SUN',
    working_days_count INT NOT NULL DEFAULT 1,
    daily_fee NUMERIC(15,2) NOT NULL DEFAULT 0,
    allowance_fee NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_budget NUMERIC(15,2) NOT NULL DEFAULT 0,
    target_sales NUMERIC(15,2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'PLANNED',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mt_spg_schedules_spg_id ON mt_spg_schedules(spg_id);
CREATE INDEX IF NOT EXISTS idx_mt_spg_schedules_key_account_id ON mt_spg_schedules(key_account_id);
CREATE INDEX IF NOT EXISTS idx_mt_spg_schedules_dates ON mt_spg_schedules(start_date, end_date);
