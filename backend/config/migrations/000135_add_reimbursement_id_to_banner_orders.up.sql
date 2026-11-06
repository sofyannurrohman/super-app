-- Add reimbursement_id column to banner_orders table
-- Backward-compatible: nullable column, old app versions still work fine.

ALTER TABLE banner_orders
    ADD COLUMN IF NOT EXISTS reimbursement_id uuid NULL;

CREATE INDEX IF NOT EXISTS idx_banner_orders_reimbursement_id
    ON banner_orders (reimbursement_id);
