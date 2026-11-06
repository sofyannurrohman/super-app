-- 000122_add_branding_banner_columns.up.sql
-- Menambahkan kolom pendukung branding banner orders di PostgreSQL

ALTER TABLE banner_orders
ADD COLUMN IF NOT EXISTS operator_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS mysql_sales_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS mysql_editor_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS mysql_updated_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS validator_id UUID REFERENCES employees(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS validator_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS validator_position VARCHAR(255),
ADD COLUMN IF NOT EXISTS purchasing_id UUID REFERENCES employees(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS purchasing_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS purchasing_position VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_banner_orders_operator_id ON banner_orders(operator_id);
CREATE INDEX IF NOT EXISTS idx_banner_orders_validator_id ON banner_orders(validator_id);
CREATE INDEX IF NOT EXISTS idx_banner_orders_purchasing_id ON banner_orders(purchasing_id);
