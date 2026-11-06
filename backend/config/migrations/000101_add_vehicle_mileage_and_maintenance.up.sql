-- 1. Add departure_mileage and return_mileage to trip tables
ALTER TABLE delivery_batches ADD COLUMN IF NOT EXISTS departure_mileage INT;
ALTER TABLE delivery_batches ADD COLUMN IF NOT EXISTS return_mileage INT;

ALTER TABLE product_transfers ADD COLUMN IF NOT EXISTS departure_mileage INT;
ALTER TABLE product_transfers ADD COLUMN IF NOT EXISTS return_mileage INT;

ALTER TABLE mt_delivery_orders ADD COLUMN IF NOT EXISTS departure_mileage INT;
ALTER TABLE mt_delivery_orders ADD COLUMN IF NOT EXISTS return_mileage INT;

ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS departure_mileage INT;
ALTER TABLE factory_returns ADD COLUMN IF NOT EXISTS return_mileage INT;

-- 2. Create vehicle_components table
CREATE TABLE IF NOT EXISTS vehicle_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    max_mileage INT NOT NULL,
    current_mileage INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create vehicle_maintenances table
CREATE TABLE IF NOT EXISTS vehicle_maintenances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    vehicle_component_id UUID NOT NULL REFERENCES vehicle_components(id) ON DELETE CASCADE,
    maintenance_date DATE NOT NULL,
    vehicle_mileage INT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
