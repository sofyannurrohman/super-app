-- 1. Tambah saldo kas kecil (petty cash) ke branches
ALTER TABLE branches ADD COLUMN IF NOT EXISTS petty_cash_balance DECIMAL(15,2) DEFAULT 0;

-- 2. Pastikan tabel cashier_financial_reports ada, lalu tambah kolom baru
CREATE TABLE IF NOT EXISTS cashier_financial_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    report_date DATE NOT NULL,
    total_cash_sales DECIMAL(15,2) DEFAULT 0,
    total_transfer_sales DECIMAL(15,2) DEFAULT 0,
    allocated_to_petty_cash DECIMAL(15,2) DEFAULT 0,
    deposited_amount DECIMAL(15,2) DEFAULT 0,
    deposit_destination_type VARCHAR(30) DEFAULT 'BRANCH_BANK',
    target_bank_account_id UUID,
    target_kas_pabrik_id UUID,
    target_bank_name VARCHAR(100),
    target_account_number VARCHAR(50),
    target_account_name VARCHAR(200),
    deposit_proof_url TEXT,
    status VARCHAR(20) DEFAULT 'DRAFT',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE cashier_financial_reports ADD COLUMN IF NOT EXISTS allocated_to_petty_cash DECIMAL(15,2) DEFAULT 0;
ALTER TABLE cashier_financial_reports ADD COLUMN IF NOT EXISTS deposit_destination_type VARCHAR(30) DEFAULT 'BRANCH_BANK';
ALTER TABLE cashier_financial_reports ADD COLUMN IF NOT EXISTS target_kas_pabrik_id UUID;

-- 3. Tambah tracking pencairan ke reimbursements
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursement_status VARCHAR(20) DEFAULT 'UNPAID';
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursed_via VARCHAR(20);
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursed_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursed_by UUID REFERENCES users(id);
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursement_proof_url TEXT;

-- 4. Pastikan tabel cash_bank_mutations ada, lalu tambah kolom baru
CREATE TABLE IF NOT EXISTS cash_bank_mutations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id),
    mutation_type VARCHAR(20) NOT NULL,
    fund_type VARCHAR(20) NOT NULL,
    bank_account_id UUID,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    notes TEXT,
    proof_url TEXT NOT NULL DEFAULT '',
    mutation_date DATE NOT NULL DEFAULT CURRENT_DATE,
    target_kas_pabrik_id UUID,
    status_setoran VARCHAR(20) DEFAULT 'PENDING',
    category VARCHAR(50),
    reimbursement_id UUID REFERENCES reimbursements(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE cash_bank_mutations ADD COLUMN IF NOT EXISTS category VARCHAR(50);
ALTER TABLE cash_bank_mutations ADD COLUMN IF NOT EXISTS reimbursement_id UUID REFERENCES reimbursements(id);
