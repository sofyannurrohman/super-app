-- Migration Down: Remove return type and product exchange features
-- Version: 000057

ALTER TABLE delivery_items DROP COLUMN IF EXISTS sales_subsequent_item_id;

DROP TABLE IF EXISTS sales_subsequent_items;
DROP TABLE IF EXISTS product_exchange_items;
DROP TABLE IF EXISTS product_exchanges;

ALTER TABLE sales_returns DROP COLUMN IF EXISTS return_type;
