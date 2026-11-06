-- Create marketplace_stocks table
CREATE TABLE IF NOT EXISTS marketplace_stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL,
    product_id UUID,
    product_name VARCHAR(200) NOT NULL DEFAULT '',
    sku VARCHAR(100) NOT NULL DEFAULT '',
    qty_unpacked INT NOT NULL DEFAULT 0,
    qty_packed INT NOT NULL DEFAULT 0,
    quantity INT NOT NULL DEFAULT 0,
    min_stock INT NOT NULL DEFAULT 0,
    price DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on branch_id for faster queries
CREATE INDEX IF NOT EXISTS idx_marketplace_stocks_branch_id ON marketplace_stocks(branch_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_stocks_product_id ON marketplace_stocks(product_id);

-- Create marketplace_opnames table
CREATE TABLE IF NOT EXISTS marketplace_opnames (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    marketplace_stock_id UUID NOT NULL REFERENCES marketplace_stocks(id) ON DELETE CASCADE,
    location VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    quantity_before INT NOT NULL DEFAULT 0,
    quantity_after INT NOT NULL DEFAULT 0,
    price DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_value DECIMAL(15,2) NOT NULL DEFAULT 0,
    reference VARCHAR(100),
    notes TEXT,
    created_by VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index on marketplace_stock_id
CREATE INDEX IF NOT EXISTS idx_marketplace_opnames_stock_id ON marketplace_opnames(marketplace_stock_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_opnames_type ON marketplace_opnames(type);
CREATE INDEX IF NOT EXISTS idx_marketplace_opnames_created_at ON marketplace_opnames(created_at);
