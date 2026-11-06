ALTER TABLE marketplace_opnames
  ADD COLUMN IF NOT EXISTS source_type VARCHAR(20),
  ADD COLUMN IF NOT EXISTS marketplace_order_id VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_marketplace_opnames_source_type ON marketplace_opnames (source_type);
CREATE INDEX IF NOT EXISTS idx_marketplace_opnames_marketplace_order_id ON marketplace_opnames (marketplace_order_id);

UPDATE marketplace_opnames
SET source_type = 'TIKTOK', marketplace_order_id = reference
WHERE type = 'STOCK_OUT' AND COALESCE(TRIM(reference), '') <> ''
  AND LOWER(notes) LIKE '%packing order%' AND COALESCE(TRIM(source_type), '') = '';
