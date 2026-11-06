DROP INDEX IF EXISTS idx_wdot_sales_bonus_request_id;

ALTER TABLE warehouse_delivery_order_transfers
DROP COLUMN sales_bonus_request_id,
ALTER COLUMN sales_transfer_id SET NOT NULL;
