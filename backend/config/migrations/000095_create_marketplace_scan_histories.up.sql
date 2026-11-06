CREATE TABLE IF NOT EXISTS marketplace_scan_histories (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(100) NOT NULL,
    action VARCHAR(20) NOT NULL,
    actor VARCHAR(200) NOT NULL,
    seller_sku VARCHAR(100) NOT NULL,
    product_name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_scan_history_order_id ON marketplace_scan_histories(order_id);
CREATE INDEX IF NOT EXISTS idx_scan_history_created_at ON marketplace_scan_histories(created_at);

