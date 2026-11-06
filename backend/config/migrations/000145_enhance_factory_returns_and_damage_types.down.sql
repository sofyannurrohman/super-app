-- 000145_enhance_factory_returns_and_damage_types.down.sql

ALTER TABLE factory_return_items
    DROP COLUMN IF EXISTS target_factory_warehouse_id,
    DROP COLUMN IF EXISTS inspection_notes,
    DROP COLUMN IF EXISTS disposition_action,
    DROP COLUMN IF EXISTS received_content_damaged_qty,
    DROP COLUMN IF EXISTS received_bottle_damaged_qty,
    DROP COLUMN IF EXISTS received_good_qty,
    DROP COLUMN IF EXISTS received_quantity,
    DROP COLUMN IF EXISTS expiry_date,
    DROP COLUMN IF EXISTS damage_reason,
    DROP COLUMN IF EXISTS damage_type,
    DROP COLUMN IF EXISTS return_type;

ALTER TABLE factory_returns
    DROP COLUMN IF EXISTS total_content_damaged_qty,
    DROP COLUMN IF EXISTS total_bottle_damaged_qty,
    DROP COLUMN IF EXISTS total_good_qty,
    DROP COLUMN IF EXISTS total_received_qty,
    DROP COLUMN IF EXISTS inspection_notes,
    DROP COLUMN IF EXISTS inspected_by_name,
    DROP COLUMN IF EXISTS receiver_name;
