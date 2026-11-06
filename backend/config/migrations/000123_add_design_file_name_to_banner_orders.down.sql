-- 000123_add_design_file_name_to_banner_orders.down.sql

ALTER TABLE banner_orders
DROP COLUMN IF EXISTS design_file_name;
