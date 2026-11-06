-- Migration 000140: Fix kas_advance_settlements foreign key cascade on kas_transactions delete

-- 1. Drop existing foreign key constraints on kas_advance_settlements if they exist
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN (
        SELECT constraint_name
        FROM information_schema.table_constraints
        WHERE table_name = 'kas_advance_settlements'
          AND constraint_type = 'FOREIGN KEY'
    ) LOOP
        EXECUTE 'ALTER TABLE kas_advance_settlements DROP CONSTRAINT IF EXISTS ' || quote_ident(r.constraint_name);
    END LOOP;
END $$;

-- 2. Add foreign key constraint with ON DELETE CASCADE
ALTER TABLE kas_advance_settlements
    ADD CONSTRAINT fk_kas_transactions_settlement
    FOREIGN KEY (kas_transaction_id)
    REFERENCES kas_transactions(id)
    ON DELETE CASCADE;

-- 3. Add foreign key constraint for settled_by if users table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        ALTER TABLE kas_advance_settlements
            ADD CONSTRAINT fk_kas_advance_settlements_settled_by
            FOREIGN KEY (settled_by)
            REFERENCES users(id)
            ON DELETE SET NULL;
    END IF;
END $$;
