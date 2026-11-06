-- 000173_add_approval_columns_to_factory_returns.down.sql

DROP INDEX IF EXISTS idx_factory_returns_approved_by_id;

ALTER TABLE factory_returns
    DROP COLUMN IF EXISTS approved_by_id,
    DROP COLUMN IF EXISTS approved_by_name,
    DROP COLUMN IF EXISTS approved_at,
    DROP COLUMN IF EXISTS rejection_reason;
