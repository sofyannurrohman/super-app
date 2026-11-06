CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_orders_delivery_order_no ON sales_orders(delivery_order_no) WHERE delivery_order_no IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_wdo_delivery_batch_id ON warehouse_delivery_orders(delivery_batch_id) WHERE delivery_batch_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_factory_returns_return_no ON factory_returns(return_no);
