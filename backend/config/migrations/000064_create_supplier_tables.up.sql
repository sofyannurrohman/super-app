-- +migrate Up

-- Tabel suppliers
CREATE TABLE IF NOT EXISTS suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(100),
    phone VARCHAR(30),
    email VARCHAR(100),
    address TEXT,
    city VARCHAR(100),
    npwp VARCHAR(50),
    bank_name VARCHAR(100),
    bank_account VARCHAR(50),
    bank_account_name VARCHAR(100),
    typical_products TEXT,
    notes TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_suppliers_company_id ON suppliers(company_id);

-- Tabel supplier_receipts (penerimaan barang dari supplier)
CREATE TABLE IF NOT EXISTS supplier_receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    receipt_no VARCHAR(100) UNIQUE,
    supplier_invoice_no VARCHAR(100),
    delivery_date DATE,
    received_date DATE,
    status VARCHAR(20) DEFAULT 'PENDING',
    proof_image_url TEXT,
    notes TEXT,
    received_by VARCHAR(100),
    total_items INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_supplier_receipts_company_id ON supplier_receipts(company_id);
CREATE INDEX IF NOT EXISTS idx_supplier_receipts_branch_id ON supplier_receipts(branch_id);
CREATE INDEX IF NOT EXISTS idx_supplier_receipts_supplier_id ON supplier_receipts(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_receipts_status ON supplier_receipts(status);

-- Tabel supplier_receipt_items (detail item per penerimaan)
CREATE TABLE IF NOT EXISTS supplier_receipt_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_receipt_id UUID NOT NULL REFERENCES supplier_receipts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    expected_qty INTEGER DEFAULT 0,
    received_qty INTEGER DEFAULT 0,
    rejected_qty INTEGER DEFAULT 0,
    unit VARCHAR(20) DEFAULT 'pcs',
    unit_price DECIMAL(15,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_supplier_receipt_items_receipt_id ON supplier_receipt_items(supplier_receipt_id);
CREATE INDEX IF NOT EXISTS idx_supplier_receipt_items_product_id ON supplier_receipt_items(product_id);

-- Tabel supplier_stocks (stok gudang yang bersumber dari supplier)
CREATE TABLE IF NOT EXISTS supplier_stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT idx_supplier_stock_branch_product_supplier UNIQUE (branch_id, product_id, supplier_id)
);

CREATE INDEX IF NOT EXISTS idx_supplier_stocks_branch_id ON supplier_stocks(branch_id);
CREATE INDEX IF NOT EXISTS idx_supplier_stocks_supplier_id ON supplier_stocks(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_stocks_product_id ON supplier_stocks(product_id);
