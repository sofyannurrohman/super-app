-- 000151_add_dual_validators_to_banner_orders.up.sql
ALTER TABLE banner_orders
ADD COLUMN IF NOT EXISTS design_validator_id UUID REFERENCES employees(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS design_validator_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS design_validator_position VARCHAR(255),
ADD COLUMN IF NOT EXISTS design_approved_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS store_validator_id UUID REFERENCES employees(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS store_validator_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS store_validator_position VARCHAR(255),
ADD COLUMN IF NOT EXISTS store_approved_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_banner_orders_design_validator_id ON banner_orders(design_validator_id);
CREATE INDEX IF NOT EXISTS idx_banner_orders_store_validator_id ON banner_orders(store_validator_id);

-- Backfill legacy records that have validator_name / validator_id and are approved
UPDATE banner_orders
SET 
  design_validator_id = COALESCE(design_validator_id, validator_id),
  design_validator_name = COALESCE(design_validator_name, validator_name),
  design_validator_position = COALESCE(design_validator_position, validator_position),
  design_approved_at = COALESCE(design_approved_at, validated_at, updated_at)
WHERE is_design_approved = true AND (design_validator_name IS NULL OR design_validator_name = '');

UPDATE banner_orders
SET 
  store_validator_id = COALESCE(store_validator_id, validator_id),
  store_validator_name = COALESCE(store_validator_name, validator_name),
  store_validator_position = COALESCE(store_validator_position, validator_position),
  store_approved_at = COALESCE(store_approved_at, validated_at, updated_at)
WHERE is_store_approved = true AND (store_validator_name IS NULL OR store_validator_name = '');
