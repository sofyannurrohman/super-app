ALTER TABLE product_transfers 
DROP COLUMN IF EXISTS transfer_type,
DROP COLUMN IF EXISTS is_billable;
