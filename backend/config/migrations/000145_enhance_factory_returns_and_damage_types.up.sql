-- 000145_enhance_factory_returns_and_damage_types.up.sql

-- Add columns to factory_returns for inspection & reception summary
ALTER TABLE factory_returns
    ADD COLUMN IF NOT EXISTS receiver_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS inspected_by_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS inspection_notes TEXT,
    ADD COLUMN IF NOT EXISTS total_received_qty INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_good_qty INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_bottle_damaged_qty INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_content_damaged_qty INT DEFAULT 0;

-- Add columns to factory_return_items for damage classification and QC breakdown
ALTER TABLE factory_return_items
    ADD COLUMN IF NOT EXISTS return_type VARCHAR(30) DEFAULT 'DAMAGED',
    ADD COLUMN IF NOT EXISTS damage_type VARCHAR(50) DEFAULT 'OTHER',
    ADD COLUMN IF NOT EXISTS damage_reason TEXT,
    ADD COLUMN IF NOT EXISTS expiry_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS received_quantity INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS received_good_qty INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS received_bottle_damaged_qty INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS received_content_damaged_qty INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS disposition_action VARCHAR(50) DEFAULT 'PENDING',
    ADD COLUMN IF NOT EXISTS inspection_notes TEXT,
    ADD COLUMN IF NOT EXISTS target_factory_warehouse_id UUID REFERENCES factory_warehouses(id) ON DELETE SET NULL;
