-- Sinkronkan faktur lama yang sudah melewati approval Admin Nota dan Kasir.
-- Status PENDING pada data lama sebelumnya membuatnya tidak tampil sebagai
-- faktur siap Penerimaan.
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
