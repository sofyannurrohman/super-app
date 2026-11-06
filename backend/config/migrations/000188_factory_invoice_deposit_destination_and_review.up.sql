ALTER TABLE cash_bank_mutations
    ADD COLUMN IF NOT EXISTS destination_company_id UUID REFERENCES companies(id),
    ADD COLUMN IF NOT EXISTS status_setoran_note TEXT,
    ADD COLUMN IF NOT EXISTS status_setoran_by UUID,
    ADD COLUMN IF NOT EXISTS status_setoran_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_cash_bank_mutations_destination_company_id
    ON cash_bank_mutations(destination_company_id);

UPDATE cash_bank_mutations mutation
SET destination_company_id = bank.company_id,
    status_setoran = 'APPROVED',
    status_setoran_at = COALESCE(mutation.status_setoran_at, mutation.created_at)
FROM company_bank_accounts bank
WHERE mutation.category = 'FACTORY_INVOICE_DEPOSIT'
  AND mutation.bank_account_id = bank.id
  AND mutation.destination_company_id IS NULL;

UPDATE cash_bank_mutations
SET status_setoran = 'APPROVED',
    status_setoran_at = COALESCE(status_setoran_at, created_at)
WHERE category = 'FACTORY_INVOICE_DEPOSIT'
  AND (status_setoran IS NULL OR status_setoran = '' OR status_setoran <> 'APPROVED');
