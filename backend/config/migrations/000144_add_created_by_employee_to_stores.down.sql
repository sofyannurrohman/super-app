DROP INDEX IF EXISTS idx_stores_created_by_employee;
ALTER TABLE stores DROP CONSTRAINT IF EXISTS fk_stores_created_by_employee;
ALTER TABLE stores DROP COLUMN IF EXISTS created_by_employee_id;
