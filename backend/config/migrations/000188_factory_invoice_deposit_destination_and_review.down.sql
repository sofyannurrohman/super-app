DROP INDEX IF EXISTS idx_cash_bank_mutations_destination_company_id;
ALTER TABLE cash_bank_mutations
    DROP COLUMN IF EXISTS status_setoran_at,
    DROP COLUMN IF EXISTS status_setoran_by,
    DROP COLUMN IF EXISTS status_setoran_note,
    DROP COLUMN IF EXISTS destination_company_id;
