-- Add soft delete columns to sales_orders and its related child tables
ALTER TABLE sales_orders
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL,
    ADD COLUMN IF NOT EXISTS deleted_by_id UUID NULL,
    ADD COLUMN IF NOT EXISTS deletion_reason TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_sales_orders_deleted_at ON sales_orders(deleted_at);

ALTER TABLE sales_order_items
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL;

CREATE INDEX IF NOT EXISTS idx_sales_order_items_deleted_at ON sales_order_items(deleted_at);

ALTER TABLE sales_order_bonus_items
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL;

CREATE INDEX IF NOT EXISTS idx_sales_order_bonus_items_deleted_at ON sales_order_bonus_items(deleted_at);

-- Add soft delete columns to sales_transactions and its related child tables
ALTER TABLE sales_transactions
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL,
    ADD COLUMN IF NOT EXISTS deleted_by_id UUID NULL,
    ADD COLUMN IF NOT EXISTS deletion_reason TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_sales_transactions_deleted_at ON sales_transactions(deleted_at);

ALTER TABLE sales_items
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL;

CREATE INDEX IF NOT EXISTS idx_sales_items_deleted_at ON sales_items(deleted_at);

ALTER TABLE sales_installments
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL;

CREATE INDEX IF NOT EXISTS idx_sales_installments_deleted_at ON sales_installments(deleted_at);

ALTER TABLE sales_payments
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL;

CREATE INDEX IF NOT EXISTS idx_sales_payments_deleted_at ON sales_payments(deleted_at);

ALTER TABLE sales_transaction_bonus_items
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL;

CREATE INDEX IF NOT EXISTS idx_sales_transaction_bonus_items_deleted_at ON sales_transaction_bonus_items(deleted_at);
