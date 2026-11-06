DROP INDEX IF EXISTS idx_marketplace_orders_return_status;
DROP INDEX IF EXISTS idx_marketplace_orders_settlement_date;
DROP INDEX IF EXISTS idx_marketplace_orders_settlement_status;

ALTER TABLE marketplace_orders
    DROP COLUMN IF EXISTS opname_id,
    DROP COLUMN IF EXISTS restocked_by,
    DROP COLUMN IF EXISTS restocked_at,
    DROP COLUMN IF EXISTS return_item_count,
    DROP COLUMN IF EXISTS return_status,
    DROP COLUMN IF EXISTS refund_reason,
    DROP COLUMN IF EXISTS refund_amount,
    DROP COLUMN IF EXISTS shipping_fee_deduction,
    DROP COLUMN IF EXISTS commission_fee,
    DROP COLUMN IF EXISTS platform_fee,
    DROP COLUMN IF EXISTS settlement_amount,
    DROP COLUMN IF EXISTS estimated_settlement_date,
    DROP COLUMN IF EXISTS settlement_date,
    DROP COLUMN IF EXISTS settlement_status;
