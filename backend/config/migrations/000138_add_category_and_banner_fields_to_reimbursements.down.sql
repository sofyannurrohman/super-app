-- Rollback category and banner installation fields from reimbursements table
DROP INDEX IF EXISTS idx_reimbursements_disbursement_status;
DROP INDEX IF EXISTS idx_reimbursements_category;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS unit_rate;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS banner_count;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS category;
