-- Migration Rollback: Remove payment_due_date from sales_orders and total_tempo_collected from delivery_batches
-- Version: 000043

ALTER TABLE sales_orders DROP COLUMN IF EXISTS payment_due_date;
ALTER TABLE delivery_batches DROP COLUMN IF EXISTS total_tempo_collected;
