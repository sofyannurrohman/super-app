CREATE TABLE IF NOT EXISTS mt_spg_daily_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spg_id UUID NOT NULL REFERENCES mt_spgs(id) ON DELETE CASCADE,
    key_account_id UUID NOT NULL REFERENCES key_accounts(id) ON DELETE CASCADE,
    supermarket_name VARCHAR(255) NOT NULL,
    report_date DATE NOT NULL,
    shift VARCHAR(100),
    total_amount DECIMAL(15, 2) DEFAULT 0,
    total_qty INTEGER DEFAULT 0,
    receipt_photo_url VARCHAR(255),
    attendance_photo_url VARCHAR(255),
    notes TEXT,
    status VARCHAR(50) DEFAULT 'SUBMITTED',
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS mt_spg_sale_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES mt_spg_daily_reports(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    product_name VARCHAR(255) NOT NULL,
    sku_code VARCHAR(100),
    unit_price DECIMAL(15, 2) DEFAULT 0,
    quantity INTEGER DEFAULT 1,
    subtotal DECIMAL(15, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_mt_spg_daily_reports_spg_id ON mt_spg_daily_reports(spg_id);
CREATE INDEX IF NOT EXISTS idx_mt_spg_daily_reports_key_account_id ON mt_spg_daily_reports(key_account_id);
CREATE INDEX IF NOT EXISTS idx_mt_spg_daily_reports_report_date ON mt_spg_daily_reports(report_date);
CREATE INDEX IF NOT EXISTS idx_mt_spg_daily_reports_status ON mt_spg_daily_reports(status);
CREATE INDEX IF NOT EXISTS idx_mt_spg_sale_items_report_id ON mt_spg_sale_items(report_id);
