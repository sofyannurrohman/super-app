ALTER TABLE marketplace_orders
    DROP COLUMN IF EXISTS fee_and_tax_amount,
    DROP COLUMN IF EXISTS shipping_cost_amount,
    DROP COLUMN IF EXISTS adjustment_amount,
    DROP COLUMN IF EXISTS fee_tax_breakdown;
