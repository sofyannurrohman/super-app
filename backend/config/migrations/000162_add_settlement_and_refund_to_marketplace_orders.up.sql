-- Add settlement and refund tracking columns to marketplace_orders
ALTER TABLE marketplace_orders
    -- Finansial & Status Pencairan
    ADD COLUMN IF NOT EXISTS settlement_status VARCHAR(50) DEFAULT 'PENDING_SETTLEMENT',
    ADD COLUMN IF NOT EXISTS settlement_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS estimated_settlement_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS settlement_amount DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS platform_fee DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS commission_fee DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS shipping_fee_deduction DECIMAL(15,2) DEFAULT 0,
    
    -- Refund & Retur Barang
    ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS refund_reason VARCHAR(255),
    ADD COLUMN IF NOT EXISTS return_status VARCHAR(50) DEFAULT 'NO_RETURN',
    ADD COLUMN IF NOT EXISTS return_item_count INT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS restocked_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS restocked_by VARCHAR(200),
    ADD COLUMN IF NOT EXISTS opname_id UUID REFERENCES marketplace_opnames(id) ON DELETE SET NULL;

-- Create indexes for quick filtering
CREATE INDEX IF NOT EXISTS idx_marketplace_orders_settlement_status ON marketplace_orders(settlement_status);
CREATE INDEX IF NOT EXISTS idx_marketplace_orders_settlement_date ON marketplace_orders(settlement_date);
CREATE INDEX IF NOT EXISTS idx_marketplace_orders_return_status ON marketplace_orders(return_status);
