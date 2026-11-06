-- 000136_add_vendor_email_tracking_to_banner_orders.down.sql
-- Rollback: hapus kolom pelacakan status email desain vendor dari banner_orders

DROP INDEX IF EXISTS idx_banner_orders_vendor_email_sent;

ALTER TABLE banner_orders
DROP COLUMN IF EXISTS vendor_email_sent_at,
DROP COLUMN IF EXISTS is_vendor_email_sent;
