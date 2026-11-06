-- Migration: Add payment, delivery, and batch fields to sales orders and delivery items
-- Version: 000038

ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS payment_collected_at TIMESTAMPTZ;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS payment_collected_amount DECIMAL(15,2) DEFAULT 0;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50);

ALTER TABLE delivery_items ALTER COLUMN sales_transaction_id DROP NOT NULL;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS sales_order_id UUID REFERENCES sales_orders(id) ON DELETE CASCADE;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS barcode_data VARCHAR(255);
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_status VARCHAR(50) DEFAULT 'UNPAID';
