-- Rollback: Enhance product_transfers for dynamic units and armada assignment
-- Version: 000035

ALTER TABLE product_transfers 
    DROP COLUMN IF EXISTS unit,
    DROP COLUMN IF EXISTS vehicle_id,
    DROP COLUMN IF EXISTS driver_id;
