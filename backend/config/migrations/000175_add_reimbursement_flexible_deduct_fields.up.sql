CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID,
    code VARCHAR(20) NOT NULL DEFAULT '',
    name VARCHAR(200) NOT NULL DEFAULT '',
    account_type VARCHAR(30) NOT NULL DEFAULT '',
    account_subtype VARCHAR(50),
    normal_balance VARCHAR(10) NOT NULL DEFAULT '',
    parent_id UUID,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS disbursement_source_type VARCHAR(50);
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS kas_pabrik_id UUID REFERENCES kas_pabrik(id) ON DELETE SET NULL;
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS expense_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL;
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS expense_account_name VARCHAR(255);
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS bank_account_id UUID REFERENCES company_bank_accounts(id) ON DELETE SET NULL;

ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS reimbursement_id UUID REFERENCES reimbursements(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_kas_transactions_reimbursement_id ON kas_transactions(reimbursement_id);
