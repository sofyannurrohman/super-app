-- 000121_add_direct_sales_support.down.sql

DROP INDEX IF EXISTS idx_sales_transactions_kas_pabrik_id;
DROP INDEX IF EXISTS idx_sales_transactions_branch_id;
DROP INDEX IF EXISTS idx_sales_transactions_buyer_employee_id;
DROP INDEX IF EXISTS idx_sales_transactions_buyer_type;
DROP INDEX IF EXISTS idx_sales_transactions_sales_channel;

ALTER TABLE sales_transactions
DROP COLUMN IF EXISTS branch_id,
DROP COLUMN IF EXISTS kas_pabrik_id,
DROP COLUMN IF EXISTS cashier_employee_id,
DROP COLUMN IF EXISTS buyer_employee_id,
DROP COLUMN IF EXISTS buyer_phone,
DROP COLUMN IF EXISTS buyer_name,
DROP COLUMN IF EXISTS buyer_type,
DROP COLUMN IF EXISTS sales_channel;
