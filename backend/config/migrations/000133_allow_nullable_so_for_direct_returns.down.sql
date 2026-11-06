DROP INDEX IF EXISTS idx_sales_returns_transaction;
ALTER TABLE sales_returns ALTER COLUMN sales_order_id SET NOT NULL;
