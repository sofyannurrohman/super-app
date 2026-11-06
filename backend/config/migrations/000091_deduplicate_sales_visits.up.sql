-- Deduplicate sales_visits recorded at the same minute for the same employee & store
-- Keep the visit record that has check_out_time, linked transactions/sales_orders, or the earliest created_at

WITH ranked_visits AS (
    SELECT 
        id,
        ROW_NUMBER() OVER (
            PARTITION BY employee_id, store_id, DATE_TRUNC('minute', check_in_time)
            ORDER BY 
                (CASE WHEN check_out_time IS NOT NULL THEN 0 ELSE 1 END),
                (CASE WHEN EXISTS (SELECT 1 FROM sales_transactions WHERE visit_id = sales_visits.id) THEN 0 ELSE 1 END),
                (CASE WHEN EXISTS (SELECT 1 FROM sales_orders WHERE visit_id = sales_visits.id) THEN 0 ELSE 1 END),
                created_at ASC,
                id ASC
        ) as rn
    FROM sales_visits
)
DELETE FROM sales_visits
WHERE id IN (
    SELECT id FROM ranked_visits WHERE rn > 1
);
