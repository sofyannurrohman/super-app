DROP INDEX IF EXISTS idx_sales_payments_penerimaan_status;
ALTER TABLE sales_payments
DROP COLUMN IF EXISTS penerimaan_notes,
DROP COLUMN IF EXISTS penerimaan_verified_by_id,
DROP COLUMN IF EXISTS penerimaan_verified_at,
DROP COLUMN IF EXISTS penerimaan_status;
