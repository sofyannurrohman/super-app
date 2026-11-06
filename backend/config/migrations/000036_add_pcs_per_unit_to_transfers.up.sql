-- Migration: Add pcs_per_unit to product_transfers for dynamic unit conversion
-- Version: 000036

ALTER TABLE product_transfers 
    ADD COLUMN IF NOT EXISTS pcs_per_unit INT NOT NULL DEFAULT 1;
