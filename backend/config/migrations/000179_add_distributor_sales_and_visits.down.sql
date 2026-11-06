DROP INDEX IF EXISTS idx_sales_orders_user_id;
DROP INDEX IF EXISTS idx_sales_orders_distributor;
ALTER TABLE sales_orders DROP COLUMN IF EXISTS is_distributor_order;
ALTER TABLE sales_orders DROP COLUMN IF EXISTS distributor_store_id;
ALTER TABLE sales_orders DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS idx_sales_transactions_user_id;
DROP INDEX IF EXISTS idx_sales_transactions_distributor;
ALTER TABLE sales_transactions DROP COLUMN IF EXISTS is_distributor_order;
ALTER TABLE sales_transactions DROP COLUMN IF EXISTS distributor_store_id;
ALTER TABLE sales_transactions DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS idx_sales_visits_user_id;
DROP INDEX IF EXISTS idx_sales_visits_distributor;
ALTER TABLE sales_visits DROP COLUMN IF EXISTS is_distributor_visit;
ALTER TABLE sales_visits DROP COLUMN IF EXISTS distributor_store_id;
ALTER TABLE sales_visits DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS idx_users_distributor_store_id;
ALTER TABLE users DROP COLUMN IF EXISTS distributor_store_id;

DROP INDEX IF EXISTS idx_stores_distributor_id;
ALTER TABLE stores DROP COLUMN IF EXISTS distributor_id;
