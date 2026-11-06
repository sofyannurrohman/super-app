-- 000151_add_dual_validators_to_banner_orders.down.sql
DROP INDEX IF EXISTS idx_banner_orders_store_validator_id;
DROP INDEX IF EXISTS idx_banner_orders_design_validator_id;

ALTER TABLE banner_orders
DROP COLUMN IF EXISTS store_approved_at,
DROP COLUMN IF EXISTS store_validator_position,
DROP COLUMN IF EXISTS store_validator_name,
DROP COLUMN IF EXISTS store_validator_id,
DROP COLUMN IF EXISTS design_approved_at,
DROP COLUMN IF EXISTS design_validator_position,
DROP COLUMN IF EXISTS design_validator_name,
DROP COLUMN IF EXISTS design_validator_id;
