ALTER TABLE kas_transactions DROP COLUMN IF EXISTS expense_account_id;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS expense_type_id;
ALTER TABLE kas_transactions DROP COLUMN IF EXISTS income_source_id;

DROP TABLE IF EXISTS kas_expense_accounts CASCADE;
DROP TABLE IF EXISTS kas_expense_types CASCADE;
DROP TABLE IF EXISTS kas_income_sources CASCADE;
