-- 000136_add_vendor_email_tracking_to_banner_orders.up.sql
-- Menambahkan kolom pelacakan status email desain vendor ke banner_orders

ALTER TABLE banner_orders
ADD COLUMN IF NOT EXISTS is_vendor_email_sent BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS vendor_email_sent_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_banner_orders_vendor_email_sent ON banner_orders(is_vendor_email_sent);
