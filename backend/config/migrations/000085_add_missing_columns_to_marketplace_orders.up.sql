-- Add 8 missing columns to marketplace_orders that were defined in the domain struct
-- but never included in the original migration 000083
ALTER TABLE marketplace_orders
    ADD COLUMN IF NOT EXISTS marketplace VARCHAR(50) DEFAULT 'tiktok',
    ADD COLUMN IF NOT EXISTS buyer_username VARCHAR(200),
    ADD COLUMN IF NOT EXISTS payment_method VARCHAR(100),
    ADD COLUMN IF NOT EXISTS total_amount DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'IDR',
    ADD COLUMN IF NOT EXISTS item_count INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS shipping_provider VARCHAR(100),
    ADD COLUMN IF NOT EXISTS create_time_raw BIGINT DEFAULT 0;

-- Fix order_id column size to match domain struct (varchar(50) -> varchar(100))
ALTER TABLE marketplace_orders
    ALTER COLUMN order_id TYPE VARCHAR(100),
    ALTER COLUMN product_name TYPE VARCHAR(500);

-- Add useful indexes
CREATE INDEX IF NOT EXISTS idx_marketplace_orders_status ON marketplace_orders(order_status);
CREATE INDEX IF NOT EXISTS idx_marketplace_orders_marketplace ON marketplace_orders(marketplace);
CREATE INDEX IF NOT EXISTS idx_marketplace_orders_order_id ON marketplace_orders(order_id);
