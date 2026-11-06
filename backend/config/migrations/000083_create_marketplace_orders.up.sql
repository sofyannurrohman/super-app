CREATE TABLE IF NOT EXISTS marketplace_orders (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(50) NOT NULL,
    order_status VARCHAR(50),
    order_substatus VARCHAR(50),
    cancelation_return_type VARCHAR(50),
    sku_id VARCHAR(50),
    seller_sku VARCHAR(50),
    product_name VARCHAR(255),
    warehouse_name VARCHAR(100),
    tracking_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
