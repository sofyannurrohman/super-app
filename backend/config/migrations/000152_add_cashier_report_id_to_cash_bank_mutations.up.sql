-- Link cash/bank mutations created from cashier financial reports.
ALTER TABLE cash_bank_mutations
    ADD COLUMN IF NOT EXISTS cashier_report_id UUID;

CREATE INDEX IF NOT EXISTS idx_cash_bank_mutations_cashier_report_id
    ON cash_bank_mutations (cashier_report_id);

