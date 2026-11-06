-- Rollback: remove reimbursement_id from banner_orders
DROP INDEX IF EXISTS idx_banner_orders_reimbursement_id;

ALTER TABLE banner_orders
    DROP COLUMN IF EXISTS reimbursement_id;
