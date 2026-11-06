ALTER TABLE marketplace_orders
    DROP COLUMN IF EXISTS marketplace,
    DROP COLUMN IF EXISTS buyer_username,
    DROP COLUMN IF EXISTS payment_method,
    DROP COLUMN IF EXISTS total_amount,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS item_count,
    DROP COLUMN IF EXISTS shipping_provider,
    DROP COLUMN IF EXISTS create_time_raw;
