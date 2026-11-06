-- 000173_add_approval_columns_to_factory_returns.up.sql

ALTER TABLE factory_returns
    ADD COLUMN IF NOT EXISTS approved_by_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS approved_by_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT;

CREATE INDEX IF NOT EXISTS idx_factory_returns_approved_by_id ON factory_returns(approved_by_id);
