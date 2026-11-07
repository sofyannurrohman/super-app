-- Sinkronkan faktur lama yang sudah melewati approval Admin Nota dan Kasir.
-- Status PENDING pada data lama sebelumnya membuatnya tidak tampil sebagai
-- faktur siap Penerimaan.
-- Fresh trial installations have not run the GORM migration for these fields.
ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) DEFAULT 'UNPAID';
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS admin_nota_confirmed_at TIMESTAMPTZ;

UPDATE sales_transactions st
SET status = 'APPROVED', updated_at = NOW()
WHERE st.status = 'PENDING'
  AND st.payment_status = 'PAID'
  AND EXISTS (
    SELECT 1
    FROM sales_orders so
    WHERE so.invoice_id = st.id
      AND so.admin_nota_confirmed_at IS NOT NULL
  );
