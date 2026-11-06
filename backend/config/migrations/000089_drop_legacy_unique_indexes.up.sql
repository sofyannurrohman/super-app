-- Drop legacy unique indexes/constraints yang menyebabkan GORM AutoMigrate crash.
DROP INDEX IF EXISTS uni_sales_orders_delivery_order_no;
ALTER TABLE warehouse_delivery_orders DROP CONSTRAINT IF EXISTS uni_warehouse_delivery_orders_delivery_batch_id;
ALTER TABLE factory_returns DROP CONSTRAINT IF EXISTS uni_factory_returns_return_no;
