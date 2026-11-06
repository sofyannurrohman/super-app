-- Migration: Add Sales Orders and Reserved Stock
-- Version: 000028

-- Sales Orders table
CREATE TABLE IF NOT EXISTS sales_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    so_number VARCHAR(50) NOT NULL UNIQUE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
    store_category VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes TEXT,
    invoice_id UUID REFERENCES sales_transactions(id) ON DELETE SET NULL,
    confirmed_at TIMESTAMPTZ,
    confirmed_by_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    converted_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    rejected_by_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    reject_notes TEXT,
    order_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sales Order Items table
CREATE TABLE IF NOT EXISTS sales_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_order_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    price DECIMAL(15,2) NOT NULL DEFAULT 0,
    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0
);

-- Add reserved_quantity to warehouse_stocks for SO reserve mechanism
ALTER TABLE warehouse_stocks ADD COLUMN IF NOT EXISTS reserved_quantity INT NOT NULL DEFAULT 0;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_sales_orders_branch ON sales_orders(branch_id);
CREATE INDEX IF NOT EXISTS idx_sales_orders_company ON sales_orders(company_id);
CREATE INDEX IF NOT EXISTS idx_sales_orders_employee ON sales_orders(employee_id);
CREATE INDEX IF NOT EXISTS idx_sales_orders_status ON sales_orders(status);
CREATE INDEX IF NOT EXISTS idx_sales_orders_store ON sales_orders(store_id);
CREATE INDEX IF NOT EXISTS idx_sales_order_items_order ON sales_order_items(sales_order_id);
