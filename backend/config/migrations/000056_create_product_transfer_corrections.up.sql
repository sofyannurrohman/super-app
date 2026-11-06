-- +migrate Up
CREATE TABLE IF NOT EXISTS product_transfer_corrections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_transfer_id UUID NOT NULL REFERENCES product_transfers(id) ON DELETE CASCADE,
    corrected_by VARCHAR(100) NOT NULL,
    old_actual_quantity INTEGER NOT NULL,
    new_actual_quantity INTEGER NOT NULL,
    old_damaged_quantity INTEGER NOT NULL,
    new_damaged_quantity INTEGER NOT NULL,
    correction_reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
