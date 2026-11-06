ALTER TABLE production_logs DROP CONSTRAINT IF EXISTS fk_production_logs_warehouse;
ALTER TABLE production_logs DROP COLUMN IF EXISTS target_warehouse_id;

ALTER TABLE factory_stocks DROP CONSTRAINT IF EXISTS fk_factory_stocks_warehouse;
ALTER TABLE factory_stocks DROP COLUMN IF EXISTS factory_warehouse_id;

DROP TABLE IF EXISTS factory_warehouses;
