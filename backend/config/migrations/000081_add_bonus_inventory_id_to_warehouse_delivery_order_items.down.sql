ALTER TABLE warehouse_delivery_order_items
DROP COLUMN bonus_inventory_id,
ALTER COLUMN product_id SET NOT NULL;
