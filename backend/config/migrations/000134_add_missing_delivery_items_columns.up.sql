-- 000134_add_missing_delivery_items_columns.up.sql
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS is_subsequent BOOLEAN DEFAULT false;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS is_collection_only BOOLEAN DEFAULT false;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS sales_transfer_id UUID;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS return_quantity INT DEFAULT 0;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS return_reason TEXT;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS has_return BOOLEAN DEFAULT false;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS signature_image_url TEXT;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS pod_image_url TEXT;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS received_by VARCHAR(100);
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_collected BOOLEAN DEFAULT false;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_amount DECIMAL(15,2) DEFAULT 0;
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50);
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS payment_collected_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_delivery_items_is_subsequent ON delivery_items(is_subsequent);
CREATE INDEX IF NOT EXISTS idx_delivery_items_sales_transfer_id ON delivery_items(sales_transfer_id);
