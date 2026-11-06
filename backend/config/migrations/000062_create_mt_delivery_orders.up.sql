-- +migrate Up
CREATE TABLE IF NOT EXISTS mt_delivery_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    do_number VARCHAR(100) NOT NULL UNIQUE,
    key_account_id UUID NOT NULL REFERENCES key_accounts(id) ON DELETE CASCADE,
    po_number VARCHAR(100),
    source_type VARCHAR(20) DEFAULT 'BRANCH',
    source_factory_id UUID,
    source_branch_id UUID,
    status VARCHAR(20) DEFAULT 'DRAFT',
    dc_destination VARCHAR(255),
    scheduled_date DATE,
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    rejection_reason TEXT,
    approved_by VARCHAR(100),
    prepared_by VARCHAR(100),
    vehicle_id UUID,
    driver_id UUID,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mt_delivery_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mt_delivery_order_id UUID NOT NULL REFERENCES mt_delivery_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    requested_qty INTEGER DEFAULT 0,
    prepared_qty INTEGER DEFAULT 0,
    shipped_qty INTEGER DEFAULT 0,
    unit VARCHAR(20) DEFAULT 'pcs',
    batch_no VARCHAR(50) DEFAULT 'DEFAULT',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
