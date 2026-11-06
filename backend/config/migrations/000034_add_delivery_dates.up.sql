ALTER TABLE product_transfers ADD COLUMN IF NOT EXISTS estimated_arrival timestamptz;
ALTER TABLE product_transfers ADD COLUMN IF NOT EXISTS target_shipment_date timestamptz;
ALTER TABLE product_transfers ADD COLUMN IF NOT EXISTS initiated_by varchar(20) DEFAULT 'FACTORY';

