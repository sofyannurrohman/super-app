CREATE TABLE IF NOT EXISTS marketplace_product_mappings (
    id SERIAL PRIMARY KEY,
    seller_sku VARCHAR(100) NOT NULL,
    marketplace_inventory_id UUID NOT NULL,
    quantity INT DEFAULT 1,
    CONSTRAINT fk_marketplace_inventory
        FOREIGN KEY (marketplace_inventory_id)
        REFERENCES marketplace_inventories(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_seller_sku ON marketplace_product_mappings(seller_sku);
