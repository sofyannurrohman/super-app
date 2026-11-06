-- 000134_add_missing_delivery_items_columns.down.sql
DROP INDEX IF EXISTS idx_delivery_items_sales_transfer_id;
DROP INDEX IF EXISTS idx_delivery_items_is_subsequent;

ALTER TABLE delivery_items DROP COLUMN IF EXISTS payment_collected_at;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS payment_method;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS payment_amount;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS payment_collected;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS received_by;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS pod_image_url;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS signature_image_url;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS has_return;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS return_reason;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS return_quantity;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS sales_transfer_id;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS is_collection_only;
ALTER TABLE delivery_items DROP COLUMN IF EXISTS is_subsequent;
