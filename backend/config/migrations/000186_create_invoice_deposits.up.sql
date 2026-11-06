-- Track branch cashier deposits allocated to sales invoices and verified by central finance.
CREATE TABLE IF NOT EXISTS invoice_deposits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id UUID NOT NULL,
    company_id UUID NOT NULL,
    bank_account_id UUID NOT NULL,
    deposit_date DATE NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    proof_url TEXT,
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    submitted_by_id UUID NOT NULL,
    reviewed_by_id UUID,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoice_deposit_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deposit_id UUID NOT NULL REFERENCES invoice_deposits(id) ON DELETE CASCADE,
    sales_transaction_id UUID NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT idx_invoice_deposit_item UNIQUE (deposit_id, sales_transaction_id)
);

CREATE INDEX IF NOT EXISTS idx_invoice_deposits_branch_id ON invoice_deposits(branch_id);
CREATE INDEX IF NOT EXISTS idx_invoice_deposits_company_id ON invoice_deposits(company_id);
CREATE INDEX IF NOT EXISTS idx_invoice_deposits_bank_account_id ON invoice_deposits(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_invoice_deposits_status ON invoice_deposits(status);
CREATE INDEX IF NOT EXISTS idx_invoice_deposit_items_sales_transaction_id ON invoice_deposit_items(sales_transaction_id);
