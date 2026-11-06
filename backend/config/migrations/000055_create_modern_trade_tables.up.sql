-- +migrate Up
CREATE TABLE IF NOT EXISTS key_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'NATIONAL',
    logo_color VARCHAR(50) DEFAULT 'bg-indigo-600',
    hq_address TEXT,
    pic_name VARCHAR(255),
    pic_phone VARCHAR(50),
    pic_email VARCHAR(255),
    total_stores INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mt_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_account_id UUID NOT NULL REFERENCES key_accounts(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    payment_term_days INTEGER DEFAULT 30,
    rebate_percent DECIMAL(5,2) DEFAULT 0,
    marketing_contribution_percent DECIMAL(5,2) DEFAULT 0,
    listing_fee_per_sku DECIMAL(15,2) DEFAULT 0,
    listed_skus_count INTEGER DEFAULT 0,
    service_level_target DECIMAL(5,2) DEFAULT 95.0,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mt_purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_account_id UUID NOT NULL REFERENCES key_accounts(id) ON DELETE CASCADE,
    po_number VARCHAR(100) NOT NULL,
    order_date DATE NOT NULL,
    delivery_deadline DATE NOT NULL,
    total_amount DECIMAL(15,2) DEFAULT 0,
    items_ordered INTEGER DEFAULT 0,
    items_shipped INTEGER DEFAULT 0,
    fill_rate DECIMAL(5,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'PENDING',
    dc_destination VARCHAR(255),
    po_schedule DATE,
    invoice_photo_url VARCHAR(255),
    receipt_photo_url VARCHAR(255),
    billing_schedule DATE,
    return_schedule DATE,
    event_schedule DATE,
    spg_schedule DATE,
    stock_update_log TEXT,
    promo_update_log TEXT,
    supermarket_name VARCHAR(255),
    md_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mt_merchandisings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_account_id UUID NOT NULL REFERENCES key_accounts(id) ON DELETE CASCADE,
    merchandiser_name VARCHAR(255) NOT NULL,
    store_name VARCHAR(255) NOT NULL,
    visit_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    planogram_compliant BOOLEAN DEFAULT TRUE,
    shelf_share_percent DECIMAL(5,2) DEFAULT 0,
    out_of_stock_skus TEXT,
    photo_url VARCHAR(255),
    status VARCHAR(50) DEFAULT 'COMPLIANT',
    invoice_photo_url VARCHAR(255),
    receipt_photo_url VARCHAR(255),
    stock_update_log TEXT,
    promo_update_log TEXT,
    supermarket_name VARCHAR(255),
    md_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mt_promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_account_id UUID NOT NULL REFERENCES key_accounts(id) ON DELETE CASCADE,
    promo_type VARCHAR(50) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    cost DECIMAL(15,2) DEFAULT 0,
    sales_target DECIMAL(15,2) DEFAULT 0,
    sales_actual DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'APPROVED',
    outlet_coverage VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
