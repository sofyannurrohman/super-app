-- 1. Add branch_id to marketplace_tokens
ALTER TABLE marketplace_tokens ADD COLUMN IF NOT EXISTS branch_id UUID REFERENCES branches(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_marketplace_tokens_branch_id ON marketplace_tokens(branch_id);

-- 2. Add branch_id to marketplace_orders
ALTER TABLE marketplace_orders ADD COLUMN IF NOT EXISTS branch_id UUID REFERENCES branches(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_marketplace_orders_branch_id ON marketplace_orders(branch_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_orders_branch_status ON marketplace_orders(branch_id, order_status);

-- 3. Add branch_id to marketplace_products
ALTER TABLE marketplace_products ADD COLUMN IF NOT EXISTS branch_id UUID REFERENCES branches(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_marketplace_products_branch_id ON marketplace_products(branch_id);

-- Drop old unique constraint on (sku_id) or (sku_id, marketplace) if exists, and recreate with branch_id
DROP INDEX IF EXISTS idx_mp_sku_marketplace;
CREATE UNIQUE INDEX IF NOT EXISTS idx_mp_sku_marketplace_branch ON marketplace_products(sku_id, marketplace, COALESCE(branch_id, '00000000-0000-0000-0000-000000000000'::uuid));

-- 4. Add branch_id to marketplace_product_mappings
ALTER TABLE marketplace_product_mappings ADD COLUMN IF NOT EXISTS branch_id UUID REFERENCES branches(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_marketplace_product_mappings_branch_id ON marketplace_product_mappings(branch_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_product_mappings_sku_branch ON marketplace_product_mappings(sku_id, branch_id);

-- 5. Add branch_id to marketplace_scan_histories
ALTER TABLE marketplace_scan_histories ADD COLUMN IF NOT EXISTS branch_id UUID REFERENCES branches(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_marketplace_scan_histories_branch_id ON marketplace_scan_histories(branch_id);

-- 6. Existing marketplace rows remain unassigned until a client maps them to a branch.
