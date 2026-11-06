-- Preserve product transfers when their driver is deleted.
-- Older installations can have both the SQL and GORM-created constraints.
ALTER TABLE product_transfers
    DROP CONSTRAINT IF EXISTS product_transfers_driver_id_fkey,
    DROP CONSTRAINT IF EXISTS fk_product_transfers_driver;

ALTER TABLE product_transfers
    ADD CONSTRAINT fk_product_transfers_driver
        FOREIGN KEY (driver_id) REFERENCES employees(id) ON DELETE SET NULL;
