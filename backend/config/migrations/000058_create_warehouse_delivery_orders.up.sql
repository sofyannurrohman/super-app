-- Migration: Create warehouse_delivery_orders and warehouse_delivery_order_items tables
-- Version: 000058

CREATE TABLE IF NOT EXISTS warehouse_delivery_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_batch_id UUID NOT NULL UNIQUE REFERENCES delivery_batches(id) ON DELETE CASCADE,
    do_number VARCHAR(50) NOT NULL UNIQUE,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS warehouse_delivery_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_delivery_order_id UUID NOT NULL REFERENCES warehouse_delivery_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    cartons_loaded INT NOT NULL DEFAULT 0,
    pcs_loaded INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_wh_do_batch ON warehouse_delivery_orders(delivery_batch_id);
