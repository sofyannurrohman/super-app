ALTER TABLE warehouse_delivery_order_transfers
ALTER COLUMN sales_transfer_id DROP NOT NULL,
ADD COLUMN sales_bonus_request_id UUID REFERENCES sales_bonus_requests(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_wdot_sales_bonus_request_id ON warehouse_delivery_order_transfers(sales_bonus_request_id);
