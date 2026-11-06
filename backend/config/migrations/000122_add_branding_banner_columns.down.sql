-- 000122_add_branding_banner_columns.down.sql
DROP INDEX IF EXISTS idx_banner_orders_purchasing_id;
DROP INDEX IF EXISTS idx_banner_orders_validator_id;
DROP INDEX IF EXISTS idx_banner_orders_operator_id;

ALTER TABLE banner_orders
DROP COLUMN IF EXISTS purchasing_position,
DROP COLUMN IF EXISTS purchasing_name,
DROP COLUMN IF EXISTS purchasing_id,
DROP COLUMN IF EXISTS validator_position,
DROP COLUMN IF EXISTS validator_name,
DROP COLUMN IF EXISTS validator_id,
DROP COLUMN IF EXISTS mysql_updated_at,
DROP COLUMN IF EXISTS mysql_editor_name,
DROP COLUMN IF EXISTS mysql_sales_name,
DROP COLUMN IF EXISTS operator_name;
