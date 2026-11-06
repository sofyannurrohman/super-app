CREATE TABLE marketplace_inventories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    sku VARCHAR(100),
    unit VARCHAR(50) NOT NULL,
    quantity INT DEFAULT 0,
    min_stock INT DEFAULT 0,
    notes TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE marketplace_inventory_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    marketplace_inventory_id UUID NOT NULL REFERENCES marketplace_inventories(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    quantity INT NOT NULL,
    quantity_before INT NOT NULL,
    quantity_after INT NOT NULL,
    reference VARCHAR(100),
    notes TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE marketplace_inventory_stock_opnames (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    marketplace_inventory_id UUID NOT NULL REFERENCES marketplace_inventories(id) ON DELETE CASCADE,
    quantity_system INT NOT NULL,
    quantity_actual INT NOT NULL,
    difference INT NOT NULL,
    notes TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mp_inv_branch ON marketplace_inventories(branch_id);
CREATE INDEX idx_mp_inv_log_inv_id ON marketplace_inventory_logs(marketplace_inventory_id);
CREATE INDEX idx_mp_inv_opname_inv_id ON marketplace_inventory_stock_opnames(marketplace_inventory_id);
