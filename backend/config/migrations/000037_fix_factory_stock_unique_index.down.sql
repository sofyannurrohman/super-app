-- Migration: Rollback factory_stocks unique index
-- Version: 000037

DROP INDEX IF EXISTS idx_factory_product_batch;
CREATE UNIQUE INDEX IF NOT EXISTS idx_factory_product ON factory_stocks(factory_id, product_id);
