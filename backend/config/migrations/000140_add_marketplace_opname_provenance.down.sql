DROP INDEX IF EXISTS idx_marketplace_opnames_source_type;
DROP INDEX IF EXISTS idx_marketplace_opnames_marketplace_order_id;
ALTER TABLE marketplace_opnames DROP COLUMN IF EXISTS source_type, DROP COLUMN IF EXISTS marketplace_order_id;
