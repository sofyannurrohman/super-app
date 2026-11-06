-- Migration: Rollback Add Dynamic Units and Batches to Sales Orders
-- Version: 000031

DROP TABLE IF EXISTS sales_order_item_batches;

ALTER TABLE warehouse_stocks 
    DROP COLUMN IF EXISTS batch_no,
    DROP COLUMN IF EXISTS expiry_date,
    DROP COLUMN IF EXISTS min_limit,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE factory_stocks 
    DROP COLUMN IF EXISTS batch_no,
    DROP COLUMN IF EXISTS expiry_date,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE production_logs 
    DROP COLUMN IF EXISTS batch_no,
    DROP COLUMN IF EXISTS expiry_date,
    DROP COLUMN IF EXISTS created_at;

ALTER TABLE warehouse_logs 
    DROP COLUMN IF EXISTS batch_no,
    DROP COLUMN IF EXISTS created_at;

ALTER TABLE factory_inventory_logs 
    DROP COLUMN IF EXISTS batch_no,
    DROP COLUMN IF EXISTS created_at;

ALTER TABLE product_transfers 
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE notifications 
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

DROP INDEX IF EXISTS idx_branch_product_type_batch;
CREATE UNIQUE INDEX IF NOT EXISTS idx_branch_product_type ON warehouse_stocks(branch_id, product_id, stock_type);

ALTER TABLE sales_order_items 
    DROP COLUMN IF EXISTS ordered_quantity,
    DROP COLUMN IF EXISTS unit,
    DROP COLUMN IF EXISTS pieces_per_unit,
    DROP COLUMN IF EXISTS reserved_quantity;
