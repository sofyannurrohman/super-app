-- Migration: Add missing payment status and midtrans fields to sales orders
-- Version: 000039

ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) DEFAULT 'UNPAID';
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS midtrans_transaction_id VARCHAR(100);
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS invoice_id UUID;
