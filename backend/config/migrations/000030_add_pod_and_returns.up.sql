-- Migration: Add POD and Sales Returns
-- Version: 000030

-- 1. Update sales_orders for POD
ALTER TABLE sales_orders 
    ADD COLUMN IF NOT EXISTS pod_image_url TEXT,
    ADD COLUMN IF NOT EXISTS received_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS received_by VARCHAR(100);

-- 2. Update warehouse_stocks for quarantine support
-- First, drop the old unique index
DROP INDEX IF EXISTS idx_branch_product;
-- Add stock_type column
ALTER TABLE warehouse_stocks 
    ADD COLUMN IF NOT EXISTS stock_type VARCHAR(20) NOT NULL DEFAULT 'GOOD';
-- Create new unique index including stock_type
CREATE UNIQUE INDEX IF NOT EXISTS idx_branch_product_type ON warehouse_stocks(branch_id, product_id, stock_type);

-- 3. Sales Returns table
CREATE TABLE IF NOT EXISTS sales_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_no VARCHAR(50) NOT NULL UNIQUE,
    sales_order_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE RESTRICT,
    transaction_id UUID REFERENCES sales_transactions(id) ON DELETE SET NULL,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes TEXT,
    approved_at TIMESTAMPTZ,
    approved_by_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Sales Return Items table
CREATE TABLE IF NOT EXISTS sales_return_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_return_id UUID NOT NULL REFERENCES sales_returns(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    price DECIMAL(15,2) NOT NULL DEFAULT 0,
    reason VARCHAR(255),
    condition VARCHAR(20) NOT NULL DEFAULT 'DAMAGED'
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sales_returns_so ON sales_returns(sales_order_id);
CREATE INDEX IF NOT EXISTS idx_sales_returns_branch ON sales_returns(branch_id);
CREATE INDEX IF NOT EXISTS idx_sales_returns_status ON sales_returns(status);
CREATE INDEX IF NOT EXISTS idx_sales_return_items_return ON sales_return_items(sales_return_id);
