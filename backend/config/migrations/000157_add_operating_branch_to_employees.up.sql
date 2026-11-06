ALTER TABLE employees ADD COLUMN IF NOT EXISTS operating_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_employees_operating_branch_id ON employees(operating_branch_id);
