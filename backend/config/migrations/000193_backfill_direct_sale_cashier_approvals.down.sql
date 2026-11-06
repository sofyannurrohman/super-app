DELETE FROM sales_payments
WHERE notes LIKE '[BACKFILL DIRECT SALE CASHIER] %'
  AND payment_status = 'PENDING';
