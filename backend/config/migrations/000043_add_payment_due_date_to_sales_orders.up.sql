-- Migration: Add payment_due_date to sales_orders and total_tempo_collected to delivery_batches
-- Version: 000043

ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS payment_due_date TIMESTAMPTZ;
ALTER TABLE delivery_batches ADD COLUMN IF NOT EXISTS total_tempo_collected DECIMAL(15,2) DEFAULT 0;
