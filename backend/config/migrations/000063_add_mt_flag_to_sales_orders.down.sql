ALTER TABLE sales_orders DROP COLUMN IF EXISTS is_mt;
ALTER TABLE sales_orders DROP COLUMN IF EXISTS factory_company_id;
ALTER TABLE mt_delivery_orders DROP COLUMN IF EXISTS sales_order_id;
DROP INDEX IF EXISTS idx_sales_orders_is_mt;
DROP INDEX IF EXISTS idx_mt_delivery_orders_sales_order_id;
