ALTER TABLE sales_transfers ADD COLUMN IF NOT EXISTS actual_quantity INTEGER DEFAULT 0;
ALTER TABLE sales_transfers ADD COLUMN IF NOT EXISTS damaged_quantity INTEGER DEFAULT 0;

CREATE TABLE IF NOT EXISTS sales_transfer_corrections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_transfer_id UUID NOT NULL REFERENCES sales_transfers(id) ON DELETE CASCADE,
    corrected_by VARCHAR(100) NOT NULL,
    old_quantity INTEGER NOT NULL,
    new_quantity INTEGER NOT NULL,
    old_damaged_quantity INTEGER NOT NULL,
    new_damaged_quantity INTEGER NOT NULL,
    correction_reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
