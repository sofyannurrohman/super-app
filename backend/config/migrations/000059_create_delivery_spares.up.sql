-- Migration: Create delivery_spares table
-- Version: 000059

CREATE TABLE IF NOT EXISTS delivery_spares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_batch_id UUID NOT NULL REFERENCES delivery_batches(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uni_delivery_spares_batch_product UNIQUE (delivery_batch_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_delivery_spares_batch ON delivery_spares(delivery_batch_id);
