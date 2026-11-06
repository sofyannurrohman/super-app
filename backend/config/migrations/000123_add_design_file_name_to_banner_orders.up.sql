-- 000123_add_design_file_name_to_banner_orders.up.sql
-- Menyimpan nama file asli desain banner yang diupload oleh editor/desainer
-- untuk keperluan cross-check validasi ukuran banner dari penamaan file

ALTER TABLE banner_orders
ADD COLUMN IF NOT EXISTS design_file_name VARCHAR(500);
