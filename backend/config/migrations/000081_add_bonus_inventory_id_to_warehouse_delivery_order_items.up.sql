ALTER TABLE warehouse_delivery_order_items
ALTER COLUMN product_id DROP NOT NULL,
ADD COLUMN bonus_inventory_id UUID REFERENCES bonus_inventories(id) ON DELETE RESTRICT;
