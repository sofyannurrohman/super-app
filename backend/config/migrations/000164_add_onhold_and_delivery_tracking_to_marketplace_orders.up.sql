-- Add on-hold details and delivery tracking columns to marketplace_orders
ALTER TABLE marketplace_orders
    ADD COLUMN IF NOT EXISTS hold_reason VARCHAR(255),
    ADD COLUMN IF NOT EXISTS estimated_settlement_notes VARCHAR(255),
    ADD COLUMN IF NOT EXISTS shipped_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_marketplace_orders_hold_reason ON marketplace_orders(hold_reason);
