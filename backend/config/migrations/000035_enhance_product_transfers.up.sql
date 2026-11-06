-- Migration: Enhance product_transfers for dynamic units and armada assignment
-- Version: 000035

ALTER TABLE product_transfers 
    ADD COLUMN IF NOT EXISTS unit VARCHAR(20) DEFAULT 'PCS',
    ADD COLUMN IF NOT EXISTS vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS driver_id UUID REFERENCES employees(id) ON DELETE SET NULL;
