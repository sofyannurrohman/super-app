-- 1. Kolom tambahan untuk kas_transactions: piutang cabang, armada kendaraan, kasbon dinas, dan integrasi faktur penjualan
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS is_branch_receivable BOOLEAN DEFAULT FALSE;
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS target_branch_id UUID;
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS vehicle_id UUID;
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS employee_id UUID;
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS advance_status VARCHAR(20) DEFAULT 'NONE';
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS sales_payment_id UUID;
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS sales_order_id UUID;

-- 2. Tabel penyelesaian kasbon / realisasi uang muka dinas
CREATE TABLE IF NOT EXISTS kas_advance_settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kas_transaction_id UUID NOT NULL,
    actual_expense_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    refund_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    shortage_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    notes TEXT,
    proof_url TEXT,
    settled_by UUID,
    settled_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kas_tx_advance_status ON kas_transactions(advance_status);
CREATE INDEX IF NOT EXISTS idx_kas_tx_target_branch ON kas_transactions(target_branch_id);
CREATE INDEX IF NOT EXISTS idx_kas_tx_employee ON kas_transactions(employee_id);
