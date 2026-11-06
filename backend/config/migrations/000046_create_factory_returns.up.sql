CREATE TABLE factory_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_no VARCHAR(100) NOT NULL UNIQUE,
    from_branch_id UUID NOT NULL REFERENCES branches(id),
    to_factory_id UUID NOT NULL REFERENCES factories(id),
    status VARCHAR(20) DEFAULT 'SHIPPED',
    notes TEXT,
    shipped_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    vehicle_id UUID REFERENCES vehicles(id),
    driver_id UUID REFERENCES employees(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE factory_return_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    factory_return_id UUID NOT NULL REFERENCES factory_returns(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    batch_no VARCHAR(50) DEFAULT 'DEFAULT',
    quantity INT NOT NULL,
    unit VARCHAR(20) DEFAULT 'PCS',
    pcs_per_unit INT DEFAULT 1
);
