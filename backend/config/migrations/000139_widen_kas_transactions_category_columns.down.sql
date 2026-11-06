-- Migration 000139 Down: Revert kas_transactions category and income source columns
ALTER TABLE kas_transactions ALTER COLUMN expense_category TYPE VARCHAR(50);
ALTER TABLE kas_transactions ALTER COLUMN income_source TYPE VARCHAR(50);
