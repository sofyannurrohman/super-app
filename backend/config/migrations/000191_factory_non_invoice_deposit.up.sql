ALTER TABLE cash_bank_mutations
    ADD COLUMN IF NOT EXISTS fund_origin VARCHAR(40) NOT NULL DEFAULT '';
