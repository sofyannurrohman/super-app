-- 1. Pulihkan total_amount HANYA pada sales_transactions yang 0 atau NULL dari tabel sales_payments
UPDATE sales_transactions st
SET total_amount = sp.total_paid,
    status = 'VERIFIED',
    verified_at = COALESCE(st.verified_at, NOW()),
    updated_at = NOW()
FROM (
    SELECT sales_transaction_id, SUM(amount) AS total_paid
    FROM sales_payments
    WHERE sales_transaction_id IS NOT NULL AND amount > 0
    GROUP BY sales_transaction_id
) sp
WHERE st.id = sp.sales_transaction_id
  AND (st.total_amount IS NULL OR st.total_amount = 0);

-- 2. Pastikan payment_status pada sales_payments berstatus SUCCESS untuk transaksi yang berstatus VERIFIED
UPDATE sales_payments
SET payment_status = 'SUCCESS',
    updated_at = NOW()
WHERE sales_transaction_id IN (
    SELECT id FROM sales_transactions WHERE status = 'VERIFIED'
) AND payment_status != 'SUCCESS';

-- 3. Update existing sales_kpis jika sudah ada barisnya
UPDATE sales_kpis k
SET achieved_omzet = sub.total_omzet,
    achieved_omzet_lama = sub.omzet_lama,
    achieved_omzet_baru = sub.omzet_baru,
    achieved_new_stores = sub.total_toko_baru,
    total_visits = sub.total_visits,
    updated_at = NOW()
FROM (
    SELECT 
        st.employee_id,
        st.period_month,
        st.period_year,
        SUM(st.total_amount) AS total_omzet,
        SUM(CASE WHEN st.store_category = 'TOKO_LAMA' THEN st.total_amount ELSE 0 END) AS omzet_lama,
        SUM(CASE WHEN st.store_category = 'TOKO_BARU' THEN st.total_amount ELSE 0 END) AS omzet_baru,
        COUNT(CASE WHEN st.store_category = 'TOKO_BARU' THEN 1 END) AS total_toko_baru,
        COUNT(DISTINCT st.visit_id) AS total_visits
    FROM sales_transactions st
    WHERE st.status = 'VERIFIED'
      AND st.period_month = 9 AND st.period_year = 2026
    GROUP BY st.employee_id, st.period_month, st.period_year
) sub
WHERE k.employee_id = sub.employee_id
  AND k.period_month = sub.period_month
  AND k.period_year = sub.period_year;
