-- Create technician_items table
CREATE TABLE IF NOT EXISTS technician_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    factory_id UUID NOT NULL REFERENCES factories(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'SPAREPART_MESIN',
    unit VARCHAR(50) NOT NULL DEFAULT 'PCS',
    quantity INT NOT NULL DEFAULT 0,
    min_limit INT DEFAULT 5,
    rack_location VARCHAR(100),
    estimated_unit_price DECIMAL(15,2) DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_technician_items_company_id ON technician_items(company_id);
CREATE INDEX IF NOT EXISTS idx_technician_items_factory_id ON technician_items(factory_id);
CREATE INDEX IF NOT EXISTS idx_technician_items_code ON technician_items(code);
CREATE INDEX IF NOT EXISTS idx_technician_items_category ON technician_items(category);

-- Create technician_stock_logs table
CREATE TABLE IF NOT EXISTS technician_stock_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    factory_id UUID NOT NULL REFERENCES factories(id) ON DELETE CASCADE,
    technician_item_id UUID NOT NULL REFERENCES technician_items(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    source VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    current_stock INT NOT NULL,
    technician_name VARCHAR(100),
    machine_or_purpose VARCHAR(255),
    reference_no VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_technician_stock_logs_company_id ON technician_stock_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_technician_stock_logs_factory_id ON technician_stock_logs(factory_id);
CREATE INDEX IF NOT EXISTS idx_technician_stock_logs_item_id ON technician_stock_logs(technician_item_id);
CREATE INDEX IF NOT EXISTS idx_technician_stock_logs_created_at ON technician_stock_logs(created_at);

-- Alter raw_material_receipts & items to support flexible supplier & technician goods
ALTER TABLE IF EXISTS raw_material_receipts ADD COLUMN IF NOT EXISTS supplier_name VARCHAR(255);
ALTER TABLE IF EXISTS raw_material_receipts ALTER COLUMN supplier_id DROP NOT NULL;

ALTER TABLE IF EXISTS raw_material_receipt_items ADD COLUMN IF NOT EXISTS technician_item_id UUID REFERENCES technician_items(id) ON DELETE SET NULL;
ALTER TABLE IF EXISTS raw_material_receipt_items ADD COLUMN IF NOT EXISTS item_type VARCHAR(50) DEFAULT 'RAW_MATERIAL';
ALTER TABLE IF EXISTS raw_material_receipt_items ALTER COLUMN product_id DROP NOT NULL;
