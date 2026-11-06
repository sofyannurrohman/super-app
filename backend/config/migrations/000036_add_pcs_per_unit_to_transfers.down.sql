-- Migration: Remove pcs_per_unit from product_transfers
-- Version: 000036

ALTER TABLE product_transfers 
    DROP COLUMN IF EXISTS pcs_per_unit;
