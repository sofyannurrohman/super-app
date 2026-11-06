-- +migrate Up

-- 1. Allow employee_id to be nullable in production_logs
ALTER TABLE production_logs ALTER COLUMN employee_id DROP NOT NULL;

-- 2. Add operator_name to production_logs for consolidated caching and fast search/display
ALTER TABLE production_logs ADD COLUMN IF NOT EXISTS operator_name TEXT;

-- 3. Create production_log_operators table for relational multiple operators
CREATE TABLE IF NOT EXISTS production_log_operators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    production_log_id UUID NOT NULL REFERENCES production_logs(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    operator_name VARCHAR(255) NOT NULL,
    is_manual BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_production_log_operators_production_log_id
    ON production_log_operators(production_log_id);

CREATE INDEX IF NOT EXISTS idx_production_log_operators_employee_id
    ON production_log_operators(employee_id);
