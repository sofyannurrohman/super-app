-- Direct sales entered by Admin Nota were left UNPAID without a cashier approval row.
-- Only restore invoices with no payment or deposit activity, so existing collections
-- cannot be queued for approval a second time.
INSERT INTO sales_payments (
    id, sales_transaction_id, employee_id, amount, payment_method,
    payment_status, payment_date, notes, is_verified, created_at, updated_at
)
SELECT gen_random_uuid(), st.id, st.employee_id, st.total_amount, st.payment_method,
       'PENDING', st.transaction_date,
       '[BACKFILL DIRECT SALE CASHIER] Penjualan Langsung Pabrik (' || st.receipt_no || ')',
       false, NOW(), NOW()
FROM sales_transactions st
WHERE st.sales_channel IN ('DIRECT_FACTORY', 'DIRECT_BRANCH', 'EMPLOYEE_SALE')
  AND st.payment_method IN ('CASH', 'TRANSFER', 'VA')
  AND st.payment_status = 'UNPAID'
  AND st.total_amount > 0
  AND COALESCE(st.paid_amount, 0) = 0
  AND st.status NOT IN ('REJECTED', 'DELETED')
  AND st.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM sales_payments sp
      WHERE sp.sales_transaction_id = st.id AND sp.deleted_at IS NULL
  )
  AND NOT EXISTS (
      SELECT 1 FROM invoice_deposit_items idi
      JOIN invoice_deposits idp ON idp.id = idi.deposit_id
      WHERE idi.sales_transaction_id = st.id AND idp.status <> 'REJECTED'
  );
