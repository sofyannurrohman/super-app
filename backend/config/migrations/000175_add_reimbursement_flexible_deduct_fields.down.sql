ALTER TABLE kas_transactions DROP COLUMN IF EXISTS reimbursement_id;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS bank_account_id;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS expense_account_name;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS expense_account_id;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS kas_pabrik_id;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS disbursement_source_type;
