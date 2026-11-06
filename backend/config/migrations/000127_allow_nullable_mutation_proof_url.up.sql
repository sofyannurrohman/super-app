-- Allow proof_url to be nullable in cash_bank_mutations
ALTER TABLE cash_bank_mutations ALTER COLUMN proof_url DROP NOT NULL;
