-- Alter sales_returns to allow NULL sales_order_id for direct factory sales returns
ALTER TABLE sales_returns ALTER COLUMN sales_order_id DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sales_returns_transaction ON sales_returns(transaction_id);
