ALTER TABLE product_transfers
    DROP CONSTRAINT IF EXISTS fk_product_transfers_driver;

ALTER TABLE product_transfers
    ADD CONSTRAINT fk_product_transfers_driver
        FOREIGN KEY (driver_id) REFERENCES employees(id);
