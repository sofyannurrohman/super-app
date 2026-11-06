-- Migration: Rollback sales transaction adjustments and verification columns
-- Version: 000050

DROP TABLE IF EXISTS sales_transaction_adjustment_logs;

ALTER TABLE sales_transactions 
    DROP COLUMN IF EXISTS verified_at,
    DROP COLUMN IF EXISTS verified_by_id;
