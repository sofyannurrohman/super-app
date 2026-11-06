DROP TABLE IF EXISTS mt_store_stock_logs;
DROP TABLE IF EXISTS mt_store_stocks;

ALTER TABLE mt_purchase_orders
DROP COLUMN IF EXISTS payment_term,
DROP COLUMN IF EXISTS due_date,
DROP COLUMN IF EXISTS billing_status,
DROP COLUMN IF EXISTS payment_proof_url,
DROP COLUMN IF EXISTS collected_at,
DROP COLUMN IF EXISTS collected_by;
