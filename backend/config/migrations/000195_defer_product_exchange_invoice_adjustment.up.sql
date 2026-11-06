ALTER TABLE sales_returns
    ADD COLUMN IF NOT EXISTS invoice_adjusted_at TIMESTAMP;

ALTER TABLE sales_return_items ADD COLUMN IF NOT EXISTS sales_item_id UUID;

-- Sebelum migrasi ini, pengajuan tukar produk langsung mengubah faktur.
UPDATE sales_returns
SET invoice_adjusted_at = created_at
WHERE return_type = 'TUKAR_PRODUK'
  AND transaction_id IS NOT NULL
  AND COALESCE(notes, '') NOT LIKE '%[DELIVERY_RETURN]%'
  AND invoice_adjusted_at IS NULL;
