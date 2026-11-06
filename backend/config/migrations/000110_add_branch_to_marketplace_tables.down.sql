DROP INDEX IF EXISTS idx_mp_sku_marketplace_branch;
CREATE UNIQUE INDEX IF NOT EXISTS idx_mp_sku_marketplace ON marketplace_products(sku_id, marketplace);

DROP INDEX IF EXISTS idx_marketplace_scan_histories_branch_id;
ALTER TABLE marketplace_scan_histories DROP COLUMN IF EXISTS branch_id;

DROP INDEX IF EXISTS idx_marketplace_product_mappings_sku_branch;
DROP INDEX IF EXISTS idx_marketplace_product_mappings_branch_id;
ALTER TABLE marketplace_product_mappings DROP COLUMN IF EXISTS branch_id;

DROP INDEX IF EXISTS idx_marketplace_products_branch_id;
ALTER TABLE marketplace_products DROP COLUMN IF EXISTS branch_id;

DROP INDEX IF EXISTS idx_marketplace_orders_branch_status;
DROP INDEX IF EXISTS idx_marketplace_orders_branch_id;
ALTER TABLE marketplace_orders DROP COLUMN IF EXISTS branch_id;

DROP INDEX IF EXISTS idx_marketplace_tokens_branch_id;
ALTER TABLE marketplace_tokens DROP COLUMN IF EXISTS branch_id;
