CREATE TABLE IF NOT EXISTS mt_spgs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_account_id UUID NOT NULL REFERENCES key_accounts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50),
    shift_schedule VARCHAR(255),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    target_sales DECIMAL(15,2) DEFAULT 0,
    actual_sales DECIMAL(15,2) DEFAULT 0,
    photo_url VARCHAR(255),
    supermarket_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
