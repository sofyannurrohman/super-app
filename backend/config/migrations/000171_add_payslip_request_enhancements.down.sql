ALTER TABLE payslip_requests 
DROP COLUMN IF EXISTS reason,
DROP COLUMN IF EXISTS payslip_data,
DROP COLUMN IF EXISTS approved_by,
DROP COLUMN IF EXISTS approved_at;
