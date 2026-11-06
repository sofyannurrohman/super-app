ALTER TABLE production_logs ADD COLUMN IF NOT EXISTS warehouse_number INT DEFAULT 12;
ALTER TABLE factory_stocks ADD COLUMN IF NOT EXISTS warehouse_number INT DEFAULT 12;
