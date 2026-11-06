-- Migration: Add performance indexes for Visit Management & Sales Filtering
-- Version: 000190

-- 1. Index for Preload("Transactions") on sales_visits and visit_id lookups
CREATE INDEX IF NOT EXISTS idx_sales_transactions_visit_id
    ON sales_transactions(visit_id) WHERE deleted_at IS NULL;

-- 2. Index for Preload("Items") on sales_transactions & GetStoreHistories LATERAL join
CREATE INDEX IF NOT EXISTS idx_sales_items_transaction_id
    ON sales_items(sales_transaction_id) WHERE deleted_at IS NULL;

-- 3. Index for NOT EXISTS subquery (so.invoice_id = sales_transactions.id)
CREATE INDEX IF NOT EXISTS idx_sales_orders_invoice_id
    ON sales_orders(invoice_id) WHERE deleted_at IS NULL;

-- 4. Index for visit date range filtering & GetStoreHistories LATERAL join per store
CREATE INDEX IF NOT EXISTS idx_sales_visits_check_in_time
    ON sales_visits(check_in_time DESC);

CREATE INDEX IF NOT EXISTS idx_sales_visits_store_check_in
    ON sales_visits(store_id, check_in_time DESC);

-- 5. Index for transaction date range filtering & GetStoreHistories LATERAL join per store
CREATE INDEX IF NOT EXISTS idx_sales_transactions_date
    ON sales_transactions(transaction_date DESC) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sales_transactions_store_date
    ON sales_transactions(store_id, transaction_date DESC) WHERE deleted_at IS NULL;

-- 6. Index for order date & employee filtering on sales_orders
CREATE INDEX IF NOT EXISTS idx_sales_orders_employee_order_date
    ON sales_orders(employee_id, order_date DESC) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sales_orders_order_date
    ON sales_orders(order_date DESC) WHERE deleted_at IS NULL;
