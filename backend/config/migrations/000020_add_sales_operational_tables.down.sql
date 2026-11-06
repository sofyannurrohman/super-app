ALTER TABLE sales_kpis DROP COLUMN IF EXISTS target_new_stores;

DROP TABLE IF EXISTS sales_transactions CASCADE;

DROP TABLE IF EXISTS stores CASCADE;
