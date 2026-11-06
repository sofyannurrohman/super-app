-- Migration: Fix factory_stocks unique index to support batches
-- Version: 000037

-- Remove old constraint that doesn't account for batch
DROP INDEX IF EXISTS idx_factory_product;

-- Create new unique index that includes batch_no
CREATE UNIQUE INDEX IF NOT EXISTS idx_factory_product_batch ON factory_stocks(factory_id, product_id, batch_no);
