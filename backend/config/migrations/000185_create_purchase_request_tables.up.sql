-- The original PR tables were only created by GORM AutoMigrate. Production
-- installations that skip AutoMigrate need the tables from versioned SQL.
CREATE TABLE IF NOT EXISTS raw_material_purchase_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID NOT NULL,
    factory_id UUID,
    pr_number VARCHAR(100) NOT NULL,
    request_date DATE NOT NULL,
    required_date DATE,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    notes TEXT,
    created_by_id UUID,
    created_by_name VARCHAR(100),
    reviewed_by_id UUID,
    reviewed_by_name VARCHAR(100),
    reviewed_at TIMESTAMPTZ,
    review_notes TEXT,
    rejection_reason TEXT,
    purchase_order_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS raw_material_purchase_request_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    purchase_request_id UUID NOT NULL,
    product_id UUID,
    item_name VARCHAR(255) NOT NULL DEFAULT '',
    factory_warehouse_id UUID,
    requested_qty INT NOT NULL DEFAULT 1,
    unit VARCHAR(20) DEFAULT 'PCS',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Also repair installations where AutoMigrate created an older PR schema.
ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS purchase_order_id UUID;
ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS reviewed_by_id UUID;
ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS reviewed_by_name VARCHAR(100);
ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;
ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS review_notes TEXT;
ALTER TABLE raw_material_purchase_requests ADD COLUMN IF NOT EXISTS rejection_reason TEXT;
ALTER TABLE raw_material_purchase_requests ALTER COLUMN factory_id DROP NOT NULL;
ALTER TABLE raw_material_purchase_request_items ADD COLUMN IF NOT EXISTS factory_warehouse_id UUID;
ALTER TABLE raw_material_purchase_request_items ALTER COLUMN product_id DROP NOT NULL;
ALTER TABLE raw_material_purchase_request_items ADD COLUMN IF NOT EXISTS item_name VARCHAR(255) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_rm_pr_company_created ON raw_material_purchase_requests (company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rm_pr_factory_id ON raw_material_purchase_requests (factory_id);
CREATE INDEX IF NOT EXISTS idx_rm_pr_number ON raw_material_purchase_requests (pr_number);
CREATE INDEX IF NOT EXISTS idx_rm_pr_po_id ON raw_material_purchase_requests (purchase_order_id);
CREATE INDEX IF NOT EXISTS idx_rm_pr_item_request_id ON raw_material_purchase_request_items (purchase_request_id);
