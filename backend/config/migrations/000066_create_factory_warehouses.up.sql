CREATE TABLE IF NOT EXISTS factory_warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    factory_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    type VARCHAR(50),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_factory_warehouses_factory FOREIGN KEY (factory_id) REFERENCES factories(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_factory_warehouses_code ON factory_warehouses(code);

ALTER TABLE factory_stocks ADD COLUMN factory_warehouse_id UUID;
ALTER TABLE factory_stocks ADD CONSTRAINT fk_factory_stocks_warehouse FOREIGN KEY (factory_warehouse_id) REFERENCES factory_warehouses(id) ON DELETE SET NULL;

ALTER TABLE production_logs ADD COLUMN target_warehouse_id UUID;
ALTER TABLE production_logs ADD CONSTRAINT fk_production_logs_warehouse FOREIGN KEY (target_warehouse_id) REFERENCES factory_warehouses(id) ON DELETE SET NULL;
