DROP INDEX IF EXISTS idx_mp_orders_status_settlement;

ALTER TABLE marketplace_orders
    DROP COLUMN IF EXISTS cancel_reason,
    DROP COLUMN IF EXISTS cancel_user,
    DROP COLUMN IF EXISTS fail_delivery_reason,
    DROP COLUMN IF EXISTS failed_delivery_shipping_fee,
    DROP COLUMN IF EXISTS cogs_amount,
    DROP COLUMN IF EXISTS packaging_cost;
