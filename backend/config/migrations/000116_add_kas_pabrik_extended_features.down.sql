DROP TABLE IF EXISTS kas_advance_settlements;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS sales_order_id;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS sales_payment_id;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS advance_status;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS employee_id;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS vehicle_id;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS target_branch_id;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS is_branch_receivable;
