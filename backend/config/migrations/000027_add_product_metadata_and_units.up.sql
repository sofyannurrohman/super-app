-- Create factories table if not exists
CREATE TABLE IF NOT EXISTS factories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    location TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create products table if not exists
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    sku VARCHAR(50) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    pcs_per_unit INT NOT NULL DEFAULT 1,
    weight DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE products DROP CONSTRAINT IF EXISTS products_sku_key CASCADE;
ALTER TABLE products DROP CONSTRAINT IF EXISTS uni_products_sku CASCADE;
DROP INDEX IF EXISTS uni_products_sku CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS uni_products_sku ON products(sku);

-- Create warehouse_stocks table if not exists
CREATE TABLE IF NOT EXISTS warehouse_stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    stock_type VARCHAR(20) NOT NULL DEFAULT 'GOOD',
    batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    expiry_date TIMESTAMPTZ,
    quantity INT NOT NULL DEFAULT 0,
    reserved_quantity INT NOT NULL DEFAULT 0,
    min_limit INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create product_transfers table if not exists
CREATE TABLE IF NOT EXISTS product_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_factory_id UUID NOT NULL REFERENCES factories(id) ON DELETE RESTRICT,
    to_branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    expiry_date TIMESTAMPTZ,
    quantity INT NOT NULL,
    total_weight DECIMAL(10,2),
    status VARCHAR(20) NOT NULL DEFAULT 'REQUESTED',
    delivery_order_no VARCHAR(100),
    notes TEXT,
    shipped_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    estimated_arrival TIMESTAMPTZ,
    target_shipment_date TIMESTAMPTZ,
    actual_received_quantity INT DEFAULT 0,
    damaged_quantity INT DEFAULT 0,
    initiated_by VARCHAR(20) DEFAULT 'FACTORY',
    unit VARCHAR(20) DEFAULT 'PCS',
    pcs_per_unit INT DEFAULT 1,
    rejection_reason TEXT,
    vehicle_id UUID,
    driver_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create factory_stocks table if not exists
CREATE TABLE IF NOT EXISTS factory_stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    factory_id UUID NOT NULL REFERENCES factories(id) ON DELETE RESTRICT,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    expiry_date TIMESTAMPTZ,
    quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create production_logs table if not exists
CREATE TABLE IF NOT EXISTS production_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    factory_id UUID NOT NULL REFERENCES factories(id) ON DELETE RESTRICT,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    expiry_date TIMESTAMPTZ,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    quantity INT NOT NULL,
    carton_count INT,
    pieces_per_carton INT,
    production_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create factory_inventory_logs table if not exists
CREATE TABLE IF NOT EXISTS factory_inventory_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    factory_id UUID NOT NULL REFERENCES factories(id) ON DELETE RESTRICT,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    type VARCHAR(20) NOT NULL,
    source VARCHAR(50),
    batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    quantity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create warehouse_logs table if not exists
CREATE TABLE IF NOT EXISTS warehouse_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    type VARCHAR(20) NOT NULL,
    source VARCHAR(50),
    batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    quantity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create notifications table if not exists
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50),
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create delivery_batches table if not exists
CREATE TABLE IF NOT EXISTS delivery_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
    branch_id UUID REFERENCES branches(id) ON DELETE RESTRICT,
    driver_id UUID,
    vehicle_id UUID,
    delivery_order_no VARCHAR(100),
    status VARCHAR(30) NOT NULL DEFAULT 'WAITING_APPROVAL',
    notes TEXT,
    barcode_data VARCHAR(255),
    admin_nota_id UUID,
    supervisor_id UUID,
    approved_at TIMESTAMPTZ,
    assigned_at TIMESTAMPTZ,
    picking_started_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    total_cash_collected DECIMAL(15,2) DEFAULT 0,
    total_transfer_collected DECIMAL(15,2) DEFAULT 0,
    cash_settled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE delivery_batches DROP CONSTRAINT IF EXISTS delivery_order_no_key CASCADE;
ALTER TABLE delivery_batches DROP CONSTRAINT IF EXISTS delivery_batches_delivery_order_no_key CASCADE;
ALTER TABLE delivery_batches DROP CONSTRAINT IF EXISTS uni_delivery_batches_delivery_order_no CASCADE;
DROP INDEX IF EXISTS uni_delivery_batches_delivery_order_no CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS uni_delivery_batches_delivery_order_no ON delivery_batches(delivery_order_no);

-- Create delivery_items table if not exists
CREATE TABLE IF NOT EXISTS delivery_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_batch_id UUID NOT NULL REFERENCES delivery_batches(id) ON DELETE CASCADE,
    sales_order_id UUID,
    sequence INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    notes TEXT,
    barcode_data VARCHAR(255),
    delivered_at TIMESTAMPTZ,
    received_by VARCHAR(100),
    pod_image_url TEXT,
    payment_collected BOOLEAN NOT NULL DEFAULT FALSE,
    payment_amount DECIMAL(15,2) DEFAULT 0,
    payment_method VARCHAR(50),
    payment_collected_at TIMESTAMPTZ,
    sales_transaction_id UUID REFERENCES sales_transactions(id) ON DELETE SET NULL
);

-- Create sales_payments table if not exists
CREATE TABLE IF NOT EXISTS sales_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_order_id UUID,
    sales_transaction_id UUID REFERENCES sales_transactions(id) ON DELETE SET NULL,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'SUCCESS',
    midtrans_transaction_id VARCHAR(100),
    payment_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    collected_by VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add new metadata and financial fields to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS category VARCHAR(100);
ALTER TABLE products ADD COLUMN IF NOT EXISTS brand VARCHAR(100);
ALTER TABLE products ADD COLUMN IF NOT EXISTS weight_unit VARCHAR(10) DEFAULT 'KG';
ALTER TABLE products ADD COLUMN IF NOT EXISTS cost_price DECIMAL(15, 2) DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS selling_price DECIMAL(15, 2) DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS specs TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS image_url TEXT;

-- Migration to ensure weight is decimal for precision
ALTER TABLE products ALTER COLUMN weight TYPE DECIMAL(10, 2);
