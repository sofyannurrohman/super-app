DO $$
BEGIN
    ALTER TABLE marketplace_product_mappings DROP CONSTRAINT IF EXISTS fk_marketplace_inventory;
    
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'marketplace_product_mappings' AND column_name = 'marketplace_inventory_id'
    ) THEN
        ALTER TABLE marketplace_product_mappings RENAME COLUMN marketplace_inventory_id TO marketplace_stock_id;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_marketplace_stock'
    ) THEN
        ALTER TABLE marketplace_product_mappings
        ADD CONSTRAINT fk_marketplace_stock
        FOREIGN KEY (marketplace_stock_id)
        REFERENCES marketplace_stocks(id)
        ON DELETE CASCADE;
    END IF;
END $$;

