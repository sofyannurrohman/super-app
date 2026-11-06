-- Migration 063: Tambah flag is_mt dan sales_order_id ke mt_delivery_orders
-- Tujuan: Sales Order yang dibuat oleh Sales MT ditandai is_mt=true
-- dan linked ke mt_delivery_orders agar omzet/fakturnya bisa dikelola pusat/pabrik.

-- 1. Tambah kolom is_mt ke sales_orders
ALTER TABLE sales_orders
    ADD COLUMN IF NOT EXISTS is_mt BOOLEAN NOT NULL DEFAULT FALSE;

-- 2. Tambah kolom sales_order_id ke mt_delivery_orders (jika belum ada)
ALTER TABLE mt_delivery_orders
    ADD COLUMN IF NOT EXISTS sales_order_id UUID REFERENCES sales_orders(id) ON DELETE SET NULL;

-- 3. Index untuk query MT orders
CREATE INDEX IF NOT EXISTS idx_sales_orders_is_mt ON sales_orders(is_mt) WHERE is_mt = TRUE;
CREATE INDEX IF NOT EXISTS idx_mt_delivery_orders_sales_order_id ON mt_delivery_orders(sales_order_id);

-- 4. Tambah kolom factory_company_id ke sales_orders
--    Ini diisi saat SO dibuat oleh Sales MT, berisi company_id pabrik/pusat
--    sehingga omzet bisa diatribusikan ke pusat bukan cabang
ALTER TABLE sales_orders
    ADD COLUMN IF NOT EXISTS factory_company_id UUID REFERENCES companies(id) ON DELETE SET NULL;
