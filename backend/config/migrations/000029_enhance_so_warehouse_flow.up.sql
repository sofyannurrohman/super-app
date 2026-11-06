-- Update sales_orders table for warehouse flow
ALTER TABLE sales_orders 
    ALTER COLUMN status TYPE VARCHAR(30),
    ADD COLUMN IF NOT EXISTS delivery_order_no VARCHAR(100),
    ADD COLUMN IF NOT EXISTS shipped_at TIMESTAMPTZ;

ALTER TABLE sales_orders DROP CONSTRAINT IF EXISTS sales_orders_delivery_order_no_key CASCADE;
ALTER TABLE sales_orders DROP CONSTRAINT IF EXISTS uni_sales_orders_delivery_order_no CASCADE;
DROP INDEX IF EXISTS uni_sales_orders_delivery_order_no CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS uni_sales_orders_delivery_order_no ON sales_orders(delivery_order_no);

-- Update sales_order_items for actual quantity tracking
ALTER TABLE sales_order_items 
    ADD COLUMN IF NOT EXISTS actual_quantity INT NOT NULL DEFAULT 0;
