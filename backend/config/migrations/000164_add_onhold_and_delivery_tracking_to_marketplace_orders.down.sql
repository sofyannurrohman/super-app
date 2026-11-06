-- Revert on-hold details and delivery tracking columns
DROP INDEX IF EXISTS idx_marketplace_orders_hold_reason;

ALTER TABLE marketplace_orders
    DROP COLUMN IF EXISTS hold_reason,
    DROP COLUMN IF EXISTS estimated_settlement_notes,
    DROP COLUMN IF EXISTS shipped_at,
    DROP COLUMN IF EXISTS delivered_at;
