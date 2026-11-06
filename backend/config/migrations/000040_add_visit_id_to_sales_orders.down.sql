-- Migration: Remove visit_id from sales_orders and drop sales_visits
-- Version: 000040

ALTER TABLE IF EXISTS sales_transactions DROP CONSTRAINT IF EXISTS fk_sales_transactions_visit;
DROP INDEX IF EXISTS idx_sales_orders_visit_id;
ALTER TABLE IF EXISTS sales_orders DROP COLUMN IF EXISTS visit_id;
DROP TABLE IF EXISTS sales_visits;
