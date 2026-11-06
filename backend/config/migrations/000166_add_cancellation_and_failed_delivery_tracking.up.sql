ALTER TABLE marketplace_orders
    ADD COLUMN IF NOT EXISTS cancel_reason VARCHAR(255),
    ADD COLUMN IF NOT EXISTS cancel_user VARCHAR(50),
    ADD COLUMN IF NOT EXISTS fail_delivery_reason VARCHAR(255),
    ADD COLUMN IF NOT EXISTS failed_delivery_shipping_fee DECIMAL(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cogs_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS packaging_cost DECIMAL(15,2) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_mp_orders_status_settlement ON marketplace_orders(order_status, settlement_status);
