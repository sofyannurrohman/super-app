-- Migration: Reset reserved stock and update legacy WAITING_STOCK sales orders
-- Version: 000168

-- Reset all reserved quantities in warehouse_stocks to 0
UPDATE warehouse_stocks SET reserved_quantity = 0 WHERE reserved_quantity != 0;

-- Reset all reserved quantities in sales_order_items to 0
UPDATE sales_order_items SET reserved_quantity = 0 WHERE reserved_quantity != 0;

-- Clear any BACKORDER batch reservations in sales_order_item_batches
DELETE FROM sales_order_item_batches WHERE batch_no = 'BACKORDER';

-- Transition existing WAITING_STOCK sales orders to CONFIRMED so they can be dispatched
UPDATE sales_orders SET status = 'CONFIRMED' WHERE status = 'WAITING_STOCK';
