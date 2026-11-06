-- Migrasi untuk memperbaiki source_type pada marketplace_opnames dan marketplace pada marketplace_orders
-- Pesanan dengan format Shopee (diawali 6 digit tanggal YYMMDD diikuti kombinasi huruf/angka kapital) yang sebelumnya tersimpan sebagai TIKTOK atau NULL

-- 1. Koreksi marketplace_orders yang berformat Shopee
UPDATE marketplace_orders
SET marketplace = 'shopee'
WHERE (marketplace IS NULL OR LOWER(marketplace) != 'shopee')
  AND order_id ~ '^[0-9]{6}[0-9A-Za-z]{6,14}$'
  AND order_id ~ '[A-Za-z]';

-- 2. Koreksi marketplace_opnames yang mereferensikan pesanan Shopee di tabel marketplace_orders
UPDATE marketplace_opnames
SET source_type = 'SHOPEE'
WHERE (source_type IS NULL OR UPPER(source_type) != 'SHOPEE')
  AND marketplace_order_id IS NOT NULL
  AND marketplace_order_id <> ''
  AND EXISTS (
      SELECT 1 FROM marketplace_orders mo
      WHERE mo.order_id = marketplace_opnames.marketplace_order_id
        AND LOWER(mo.marketplace) = 'shopee'
  );

-- 3. Koreksi marketplace_opnames berdasarkan format order ID Shopee pada reference atau marketplace_order_id
UPDATE marketplace_opnames
SET source_type = 'SHOPEE'
WHERE (source_type IS NULL OR UPPER(source_type) != 'SHOPEE')
  AND (
      (marketplace_order_id ~ '^[0-9]{6}[0-9A-Za-z]{6,14}$' AND marketplace_order_id ~ '[A-Za-z]')
      OR (reference ~ '^[0-9]{6}[0-9A-Za-z]{6,14}$' AND reference ~ '[A-Za-z]')
      OR (reference ~ '^RETUR-[0-9]{6}[0-9A-Za-z]{6,14}$' AND reference ~ '[A-Za-z]')
  );
