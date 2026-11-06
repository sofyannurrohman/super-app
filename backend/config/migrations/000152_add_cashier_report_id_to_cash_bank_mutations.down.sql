DROP INDEX IF EXISTS idx_cash_bank_mutations_cashier_report_id;

ALTER TABLE cash_bank_mutations
    DROP COLUMN IF EXISTS cashier_report_id;
