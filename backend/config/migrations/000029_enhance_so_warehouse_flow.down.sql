-- Rollback updates for warehouse flow
ALTER TABLE sales_order_items 
    DROP COLUMN IF EXISTS actual_quantity;

ALTER TABLE sales_orders 
    DROP COLUMN IF EXISTS shipped_at,
    DROP COLUMN IF EXISTS delivery_order_no;
