-- Rollback: Remove Sales Orders and Reserved Stock
-- Version: 000028

ALTER TABLE warehouse_stocks DROP COLUMN IF EXISTS reserved_quantity;

DROP INDEX IF EXISTS idx_sales_order_items_order;
DROP INDEX IF EXISTS idx_sales_orders_store;
DROP INDEX IF EXISTS idx_sales_orders_status;
DROP INDEX IF EXISTS idx_sales_orders_employee;
DROP INDEX IF EXISTS idx_sales_orders_company;
DROP INDEX IF EXISTS idx_sales_orders_branch;

DROP TABLE IF EXISTS sales_order_items;
DROP TABLE IF EXISTS sales_orders;
