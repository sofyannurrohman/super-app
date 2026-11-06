-- Convert empty string client_ids to NULL to prevent unique constraint conflicts
UPDATE sales_transactions SET client_id = NULL WHERE client_id = '';
UPDATE sales_orders SET client_id = NULL WHERE client_id = '';
UPDATE sales_visits SET client_id = NULL WHERE client_id = '';

-- Fix any invalid/zero company_id in sales_transactions using employee/store fallback or default company
UPDATE sales_transactions st
SET company_id = e.company_id
FROM employees e
WHERE (st.company_id = '00000000-0000-0000-0000-000000000000' OR st.company_id IS NULL)
  AND st.employee_id = e.id
  AND e.company_id IS NOT NULL;

UPDATE sales_transactions st
SET company_id = s.company_id
FROM stores s
WHERE (st.company_id = '00000000-0000-0000-0000-000000000000' OR st.company_id IS NULL)
  AND st.store_id = s.id
  AND s.company_id IS NOT NULL;

UPDATE sales_transactions
SET company_id = (SELECT id FROM companies ORDER BY created_at ASC LIMIT 1)
WHERE (company_id = '00000000-0000-0000-0000-000000000000' OR company_id IS NULL)
  AND (SELECT COUNT(*) FROM companies) > 0;
