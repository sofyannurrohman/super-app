-- Migration: Tambah kolom branding_shop_id ke tabel stores sebagai penanda sinkronisasi
-- Ini dipakai oleh CronJob Sinkronisasi untuk mendeteksi toko mana yang berasal dari Branding App
-- sehingga tidak terjadi duplikasi saat cronjob berjalan berulang kali.

ALTER TABLE stores
    ADD COLUMN IF NOT EXISTS branding_shop_id INTEGER DEFAULT NULL;

-- Index untuk mempercepat lookup saat cronjob melakukan pengecekan "sudah ada atau belum"
CREATE UNIQUE INDEX IF NOT EXISTS idx_stores_branding_shop_id
    ON stores (branding_shop_id)
    WHERE branding_shop_id IS NOT NULL;

COMMENT ON COLUMN stores.branding_shop_id IS
    'ID toko dari database MySQL aplikasi katalog eksternal. NULL = toko murni dari Super App. '
    'Digunakan oleh cronjob sinkronisasi untuk deteksi duplikasi.';
