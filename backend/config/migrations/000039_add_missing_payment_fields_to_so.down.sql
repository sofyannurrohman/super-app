-- Migration: Revert adding missing payment status and midtrans fields to sales orders
-- Version: 000039

ALTER TABLE sales_orders DROP COLUMN IF EXISTS payment_status;
ALTER TABLE sales_orders DROP COLUMN IF EXISTS midtrans_transaction_id;
