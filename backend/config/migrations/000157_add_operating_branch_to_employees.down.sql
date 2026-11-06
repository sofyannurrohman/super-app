DROP INDEX IF EXISTS idx_employees_operating_branch_id;
ALTER TABLE employees DROP COLUMN IF EXISTS operating_branch_id;
