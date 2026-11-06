-- Remove duplicate rows first (keep the lowest id per order_id+sku_id pair)
DELETE FROM marketplace_orders a
USING marketplace_orders b
WHERE a.id > b.id
  AND a.order_id = b.order_id
  AND a.sku_id = b.sku_id;

-- Add unique composite index on (order_id, sku_id)
CREATE UNIQUE INDEX IF NOT EXISTS idx_mp_orders_order_sku
    ON marketplace_orders(order_id, sku_id);
