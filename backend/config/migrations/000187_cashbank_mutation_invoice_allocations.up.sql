CREATE TABLE IF NOT EXISTS cash_bank_mutation_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mutation_id UUID NOT NULL REFERENCES cash_bank_mutations(id) ON DELETE CASCADE,
    sales_payment_id UUID NOT NULL REFERENCES sales_payments(id),
    sales_transaction_id UUID NOT NULL REFERENCES sales_transactions(id),
    amount DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cbm_invoice_mutation_id ON cash_bank_mutation_invoices(mutation_id);
CREATE INDEX IF NOT EXISTS idx_cbm_invoice_payment_id ON cash_bank_mutation_invoices(sales_payment_id);
CREATE INDEX IF NOT EXISTS idx_cbm_invoice_transaction_id ON cash_bank_mutation_invoices(sales_transaction_id);
