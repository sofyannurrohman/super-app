-- Migration Down: Remove warehouse_delivery_orders and warehouse_delivery_order_items tables
-- Version: 000058

DROP TABLE IF EXISTS warehouse_delivery_order_items;
DROP TABLE IF EXISTS warehouse_delivery_orders;
