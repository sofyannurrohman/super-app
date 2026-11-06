DROP INDEX IF EXISTS idx_sales_transaction_bonus_items_deleted_at;
ALTER TABLE sales_transaction_bonus_items DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_sales_payments_deleted_at;
ALTER TABLE sales_payments DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_sales_installments_deleted_at;
ALTER TABLE sales_installments DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_sales_items_deleted_at;
ALTER TABLE sales_items DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_sales_transactions_deleted_at;
ALTER TABLE sales_transactions
    DROP COLUMN IF EXISTS deletion_reason,
    DROP COLUMN IF EXISTS deleted_by_id,
    DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_sales_order_bonus_items_deleted_at;
ALTER TABLE sales_order_bonus_items DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_sales_order_items_deleted_at;
ALTER TABLE sales_order_items DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_sales_orders_deleted_at;
ALTER TABLE sales_orders
    DROP COLUMN IF EXISTS deletion_reason,
    DROP COLUMN IF EXISTS deleted_by_id,
    DROP COLUMN IF EXISTS deleted_at;
