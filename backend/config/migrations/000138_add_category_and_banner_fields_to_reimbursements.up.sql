-- Add category and banner installation fields to reimbursements table
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS category VARCHAR(50) DEFAULT 'GENERAL';
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS banner_count INT DEFAULT 0;
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS unit_rate DECIMAL(15,2) DEFAULT 0;

-- Indexes on category and disbursement_status for faster filtering & joins
CREATE INDEX IF NOT EXISTS idx_reimbursements_category ON reimbursements(category);
CREATE INDEX IF NOT EXISTS idx_reimbursements_disbursement_status ON reimbursements(disbursement_status);
