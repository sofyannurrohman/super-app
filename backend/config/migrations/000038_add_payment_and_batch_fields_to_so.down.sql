-- Down Migration: Remove payment, delivery, and batch fields
-- Version: 000038

ALTER TABLE delivery_items DROP COLUMN IF EXISTS payment_status;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS barcode_data;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS sales_order_id;
-- Note: cannot safely restore sales_transaction_id NOT NULL constraint if data exists

ALTER TABLE sales_orders DROP COLUMN IF EXISTS payment_method;
ALTER TABLE sales_orders DROP COLUMN IF EXISTS payment_collected_amount;
ALTER TABLE sales_orders DROP COLUMN IF EXISTS payment_collected_at;
