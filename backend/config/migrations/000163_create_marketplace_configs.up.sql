-- Create marketplace_configs table for dynamic API credentials (Shopee, TikTok, etc.)
CREATE TABLE IF NOT EXISTS marketplace_configs (
    id SERIAL PRIMARY KEY,
    marketplace VARCHAR(50) NOT NULL UNIQUE, -- 'shopee', 'tiktok', 'tokopedia'
    app_id VARCHAR(100),                     -- partner_id / app_key
    app_secret TEXT,                         -- partner_key / app_secret
    environment VARCHAR(50) DEFAULT 'production', -- 'production', 'sandbox'
    base_url TEXT,
    redirect_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_marketplace_configs_marketplace ON marketplace_configs(marketplace);
