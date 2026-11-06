-- 000121_add_direct_sales_support.up.sql

ALTER TABLE sales_transactions
ALTER COLUMN store_id DROP NOT NULL,
ADD COLUMN IF NOT EXISTS sales_channel VARCHAR(50) DEFAULT 'CANVAS',
ADD COLUMN IF NOT EXISTS buyer_type VARCHAR(50) DEFAULT 'REGISTERED_STORE',
ADD COLUMN IF NOT EXISTS buyer_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS buyer_phone VARCHAR(50),
ADD COLUMN IF NOT EXISTS buyer_employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS cashier_employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS kas_pabrik_id UUID REFERENCES kas_pabrik(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS branch_id UUID REFERENCES branches(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_sales_transactions_sales_channel ON sales_transactions(sales_channel);
CREATE INDEX IF NOT EXISTS idx_sales_transactions_buyer_type ON sales_transactions(buyer_type);
CREATE INDEX IF NOT EXISTS idx_sales_transactions_buyer_employee_id ON sales_transactions(buyer_employee_id);
CREATE INDEX IF NOT EXISTS idx_sales_transactions_branch_id ON sales_transactions(branch_id);
CREATE INDEX IF NOT EXISTS idx_sales_transactions_kas_pabrik_id ON sales_transactions(kas_pabrik_id);
