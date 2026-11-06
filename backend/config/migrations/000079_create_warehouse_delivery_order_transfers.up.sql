-- Migration Up: Create warehouse_delivery_order_transfers table

CREATE TABLE IF NOT EXISTS sales_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    product_id UUID,
    bonus_inventory_id UUID,
    quantity INTEGER NOT NULL,
    unit VARCHAR(20) DEFAULT 'pcs',
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    reference_no VARCHAR(50),
    actual_quantity INTEGER DEFAULT 0,
    damaged_quantity INTEGER DEFAULT 0,
    notes TEXT,
    transfer_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS warehouse_delivery_order_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_delivery_order_id UUID NOT NULL REFERENCES warehouse_delivery_orders(id) ON DELETE CASCADE,
    sales_transfer_id UUID NOT NULL REFERENCES sales_transfers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_wdot_warehouse_delivery_order_id ON warehouse_delivery_order_transfers(warehouse_delivery_order_id);
CREATE INDEX IF NOT EXISTS idx_wdot_sales_transfer_id ON warehouse_delivery_order_transfers(sales_transfer_id);

