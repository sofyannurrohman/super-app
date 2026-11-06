ALTER TABLE stores
    ADD COLUMN IF NOT EXISTS created_by_employee_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_stores_created_by_employee'
    ) THEN
        ALTER TABLE stores
            ADD CONSTRAINT fk_stores_created_by_employee
            FOREIGN KEY (created_by_employee_id)
            REFERENCES employees(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_stores_created_by_employee
    ON stores(created_by_employee_id);
