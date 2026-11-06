CREATE TABLE IF NOT EXISTS mt_inventories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'PROMO_GIFT',
    quantity INT NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL DEFAULT 'PCS',
    min_stock_alert INT NOT NULL DEFAULT 10,
    photo_url VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mt_inventories_company ON mt_inventories(company_id);
CREATE INDEX IF NOT EXISTS idx_mt_inventories_branch ON mt_inventories(branch_id);

CREATE TABLE IF NOT EXISTS mt_inventory_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mt_inventory_id UUID NOT NULL REFERENCES mt_inventories(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- IN, OUT, ADJUSTMENT, RETURN
    quantity INT NOT NULL,
    quantity_before INT NOT NULL DEFAULT 0,
    quantity_after INT NOT NULL DEFAULT 0,
    reference_no VARCHAR(100),
    recipient_type VARCHAR(50), -- SPG, OUTLET_MT, EVENT, GUDANG, LAINNYA
    recipient_name VARCHAR(255),
    notes TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mt_inventory_logs_inventory ON mt_inventory_logs(mt_inventory_id);
CREATE INDEX IF NOT EXISTS idx_mt_inventory_logs_type ON mt_inventory_logs(type);
