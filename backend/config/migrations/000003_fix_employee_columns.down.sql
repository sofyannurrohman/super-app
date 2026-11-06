ALTER TABLE employees 
DROP COLUMN IF EXISTS emergency_contact,
DROP COLUMN IF EXISTS bank_name,
DROP COLUMN IF EXISTS bank_account_number,
DROP COLUMN IF EXISTS account_holder_name;
