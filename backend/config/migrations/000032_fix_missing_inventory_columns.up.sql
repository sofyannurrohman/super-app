-- Migration: Add Dynamic Units and Batches to Sales Orders
-- Version: 000031

-- 1. Update sales_order_items for dynamic units and reservation tracking
ALTER TABLE sales_order_items 
    ADD COLUMN IF NOT EXISTS ordered_quantity INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS unit VARCHAR(20) NOT NULL DEFAULT 'PCS',
    ADD COLUMN IF NOT EXISTS pieces_per_unit INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS reserved_quantity INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 2. Update warehouse_stocks for batch and expiry tracking
ALTER TABLE warehouse_stocks 
    ADD COLUMN IF NOT EXISTS batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    ADD COLUMN IF NOT EXISTS expiry_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS min_limit INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 3. Update factory_stocks for batch and expiry tracking
ALTER TABLE factory_stocks 
    ADD COLUMN IF NOT EXISTS batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    ADD COLUMN IF NOT EXISTS expiry_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 4. Update production_logs for batch and expiry tracking
ALTER TABLE production_logs 
    ADD COLUMN IF NOT EXISTS batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    ADD COLUMN IF NOT EXISTS expiry_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 5. Update inventory logs for batch tracking
ALTER TABLE warehouse_logs 
    ADD COLUMN IF NOT EXISTS batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE factory_inventory_logs 
    ADD COLUMN IF NOT EXISTS batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 6. Update other AutoMigrate tables used in Ordering
ALTER TABLE product_transfers 
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE notifications 
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- 7. Create sales_order_item_batches table and ensure audit columns
CREATE TABLE IF NOT EXISTS sales_order_item_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_order_item_id UUID NOT NULL REFERENCES sales_order_items(id) ON DELETE CASCADE,
    batch_no VARCHAR(50) NOT NULL,
    quantity INT NOT NULL
);

ALTER TABLE sales_order_item_batches 
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Indexes for performance
DROP INDEX IF EXISTS idx_branch_product_type;
CREATE UNIQUE INDEX IF NOT EXISTS idx_branch_product_type_batch ON warehouse_stocks(branch_id, product_id, stock_type, batch_no);
CREATE INDEX IF NOT EXISTS idx_so_item_batches_item ON sales_order_item_batches(sales_order_item_id);
CREATE INDEX IF NOT EXISTS idx_so_item_batches_batch ON sales_order_item_batches(batch_no);
