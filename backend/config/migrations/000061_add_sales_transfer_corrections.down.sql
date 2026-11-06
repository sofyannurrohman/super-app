DROP TABLE IF EXISTS sales_transfer_corrections;
ALTER TABLE sales_transfers DROP COLUMN IF EXISTS actual_quantity;
ALTER TABLE sales_transfers DROP COLUMN IF EXISTS damaged_quantity;
