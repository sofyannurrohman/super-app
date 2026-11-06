-- Migration 000179: Add distributor secondary sales and visits tracking

-- 1. Add distributor_id to stores (for assigned retail stores)
ALTER TABLE stores ADD COLUMN IF NOT EXISTS distributor_id UUID REFERENCES stores(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_stores_distributor_id ON stores(distributor_id);

-- 2. Add distributor_store_id to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_users_distributor_store_id ON users(distributor_store_id);

-- 3. Add distributor and user fields to sales_visits
ALTER TABLE sales_visits ALTER COLUMN employee_id DROP NOT NULL;
ALTER TABLE sales_visits ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE sales_visits ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL;
ALTER TABLE sales_visits ADD COLUMN IF NOT EXISTS is_distributor_visit BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_sales_visits_distributor ON sales_visits(distributor_store_id, is_distributor_visit);
CREATE INDEX IF NOT EXISTS idx_sales_visits_user_id ON sales_visits(user_id);

-- 4. Add distributor and user fields to sales_transactions
ALTER TABLE sales_transactions ALTER COLUMN employee_id DROP NOT NULL;
ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL;
ALTER TABLE sales_transactions ADD COLUMN IF NOT EXISTS is_distributor_order BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_sales_transactions_distributor ON sales_transactions(distributor_store_id, is_distributor_order);
CREATE INDEX IF NOT EXISTS idx_sales_transactions_user_id ON sales_transactions(user_id);

-- 5. Add distributor and user fields to sales_orders
ALTER TABLE sales_orders ALTER COLUMN employee_id DROP NOT NULL;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS distributor_store_id UUID REFERENCES stores(id) ON DELETE SET NULL;
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS is_distributor_order BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_sales_orders_distributor ON sales_orders(distributor_store_id, is_distributor_order);
CREATE INDEX IF NOT EXISTS idx_sales_orders_user_id ON sales_orders(user_id);
