ALTER TABLE marketplace_product_mappings
DROP CONSTRAINT IF EXISTS fk_marketplace_stock;

ALTER TABLE marketplace_product_mappings
RENAME COLUMN marketplace_stock_id TO marketplace_inventory_id;

ALTER TABLE marketplace_product_mappings
ADD CONSTRAINT fk_marketplace_inventory
FOREIGN KEY (marketplace_inventory_id)
REFERENCES marketplace_inventories(id)
ON DELETE CASCADE;
