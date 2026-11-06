-- Drop vehicle_maintenances table
DROP TABLE IF EXISTS vehicle_maintenances CASCADE;

-- Drop vehicle_components table
DROP TABLE IF EXISTS vehicle_components CASCADE;

-- Remove mileage columns from trip tables
ALTER TABLE factory_returns DROP COLUMN IF EXISTS departure_mileage;
ALTER TABLE factory_returns DROP COLUMN IF EXISTS return_mileage;

ALTER TABLE mt_delivery_orders DROP COLUMN IF EXISTS departure_mileage;
ALTER TABLE mt_delivery_orders DROP COLUMN IF EXISTS return_mileage;

ALTER TABLE product_transfers DROP COLUMN IF EXISTS departure_mileage;
ALTER TABLE product_transfers DROP COLUMN IF EXISTS return_mileage;

ALTER TABLE delivery_batches DROP COLUMN IF EXISTS departure_mileage;
ALTER TABLE delivery_batches DROP COLUMN IF EXISTS return_mileage;
