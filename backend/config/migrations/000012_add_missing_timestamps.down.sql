-- Remove updated_at from branches
ALTER TABLE branches DROP COLUMN IF EXISTS updated_at;

-- Remove updated_at from users
ALTER TABLE users DROP COLUMN IF EXISTS updated_at;
