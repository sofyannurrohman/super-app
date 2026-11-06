CREATE TABLE IF NOT EXISTS marketplace_products (
    id SERIAL PRIMARY KEY,
    marketplace VARCHAR(50) NOT NULL DEFAULT 'tiktok',
    product_id VARCHAR(100),
    product_name VARCHAR(500),
    sku_id VARCHAR(100),
    seller_sku VARCHAR(100),
    price_amount DECIMAL(15,2) DEFAULT 0,
    price_currency VARCHAR(10) DEFAULT 'IDR',
    stock INT DEFAULT 0,
    status VARCHAR(50),
    category VARCHAR(200),
    sales_regions VARCHAR(200),
    last_synced_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_mp_sku_marketplace ON marketplace_products(sku_id, marketplace);
CREATE INDEX IF NOT EXISTS idx_mp_product_id ON marketplace_products(product_id);
