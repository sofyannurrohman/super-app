-- Tabel Antrean Produk Auto Boost Shopee
CREATE TABLE IF NOT EXISTS shopee_auto_boost_items (
    id BIGSERIAL PRIMARY KEY,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    shop_id VARCHAR(100) NOT NULL,
    item_id BIGINT NOT NULL,
    item_name VARCHAR(500) NOT NULL,
    item_sku VARCHAR(100),
    image_url TEXT,
    price_amount NUMERIC(15, 2) DEFAULT 0,
    sequence INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_boosted_at TIMESTAMPTZ,
    cooldown_until TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'IDLE',
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_shopee_auto_boost_shop_item UNIQUE (shop_id, item_id)
);

CREATE INDEX IF NOT EXISTS idx_shopee_boost_shop_active ON shopee_auto_boost_items(shop_id, is_active);
CREATE INDEX IF NOT EXISTS idx_shopee_boost_cooldown ON shopee_auto_boost_items(cooldown_until);

-- Tabel Pengaturan Auto Boost per Toko/Cabang
CREATE TABLE IF NOT EXISTS shopee_auto_boost_settings (
    id BIGSERIAL PRIMARY KEY,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    shop_id VARCHAR(100) NOT NULL UNIQUE,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    rotation_mode VARCHAR(50) NOT NULL DEFAULT 'ROUND_ROBIN',
    check_interval_minutes INT NOT NULL DEFAULT 5,
    last_run_at TIMESTAMPTZ,
    last_run_status VARCHAR(50),
    last_run_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabel Log Riwayat Boost
CREATE TABLE IF NOT EXISTS shopee_auto_boost_logs (
    id BIGSERIAL PRIMARY KEY,
    shop_id VARCHAR(100) NOT NULL,
    item_id BIGINT NOT NULL,
    item_name VARCHAR(500),
    triggered_by VARCHAR(50) NOT NULL DEFAULT 'WORKER',
    status VARCHAR(50) NOT NULL,
    message TEXT,
    boosted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shopee_boost_logs_shop ON shopee_auto_boost_logs(shop_id, boosted_at DESC);
