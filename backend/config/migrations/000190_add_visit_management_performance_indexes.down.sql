-- Rollback: Remove Visit Management performance indexes
-- Version: 000190

DROP INDEX IF EXISTS idx_sales_orders_order_date;
DROP INDEX IF EXISTS idx_sales_orders_employee_order_date;
DROP INDEX IF EXISTS idx_sales_transactions_store_date;
DROP INDEX IF EXISTS idx_sales_transactions_date;
DROP INDEX IF EXISTS idx_sales_visits_store_check_in;
DROP INDEX IF EXISTS idx_sales_visits_check_in_time;
DROP INDEX IF EXISTS idx_sales_orders_invoice_id;
DROP INDEX IF EXISTS idx_sales_items_transaction_id;
DROP INDEX IF EXISTS idx_sales_transactions_visit_id;
