ALTER TABLE raw_material_receipt_items DROP COLUMN IF EXISTS technician_item_id;
ALTER TABLE raw_material_receipt_items DROP COLUMN IF EXISTS item_type;
ALTER TABLE raw_material_receipts DROP COLUMN IF EXISTS supplier_name;

DROP TABLE IF EXISTS technician_stock_logs;
DROP TABLE IF EXISTS technician_items;
