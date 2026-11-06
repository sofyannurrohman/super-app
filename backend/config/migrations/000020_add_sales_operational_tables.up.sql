CREATE TABLE stores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    assigned_employee_id UUID REFERENCES employees(id) ON DELETE SET NULL, -- PENTING: Untuk Rayonisasi (1 Toko = 1 Sales)
    name VARCHAR(255) NOT NULL,
    owner_name VARCHAR(255),
    phone_number VARCHAR(50),
    address TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    is_active BOOLEAN DEFAULT TRUE,
    first_transaction_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE sales_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    visit_id UUID, -- PENTING: Mengikat transaksi ke log check-in GPS sales
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    store_id UUID REFERENCES stores(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    receipt_no VARCHAR(100), -- Nomor identifikasi nota
    receipt_image_url TEXT, -- PENTING: Bukti foto untuk admin
    total_amount DECIMAL(15, 2) NOT NULL,
    store_category VARCHAR(50), -- 'TOKO_LAMA', 'TOKO_BARU'
    period_month INT NOT NULL,
    period_year INT NOT NULL,
    transaction_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING', -- 'PENDING', 'VERIFIED'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE sales_kpis ADD COLUMN target_new_stores INT DEFAULT 0;
