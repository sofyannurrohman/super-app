DO $$
BEGIN
    IF to_regclass('raw_material_purchase_requests') IS NOT NULL THEN
        ALTER TABLE raw_material_purchase_requests ALTER COLUMN factory_id DROP NOT NULL;
    END IF;
    IF to_regclass('raw_material_purchase_request_items') IS NOT NULL THEN
        ALTER TABLE raw_material_purchase_request_items ALTER COLUMN product_id DROP NOT NULL;
        ALTER TABLE raw_material_purchase_request_items ADD COLUMN IF NOT EXISTS item_name VARCHAR(255) NOT NULL DEFAULT '';
    END IF;
END $$;
