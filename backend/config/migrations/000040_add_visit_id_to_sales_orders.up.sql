-- Migration: Add sales_visits table and visit_id to sales_orders
-- Version: 000040

CREATE TABLE IF NOT EXISTS sales_visits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    check_in_time TIMESTAMPTZ NOT NULL,
    check_out_time TIMESTAMPTZ,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    selfie_url TEXT,
    notes TEXT,
    type VARCHAR(20) DEFAULT 'CHECKIN',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS visit_id UUID REFERENCES sales_visits(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_sales_orders_visit_id ON sales_orders(visit_id);

-- Also add foreign key to sales_transactions if missing
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='sales_transactions' AND column_name='visit_id') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name='fk_sales_transactions_visit') THEN
            ALTER TABLE sales_transactions ADD CONSTRAINT fk_sales_transactions_visit FOREIGN KEY (visit_id) REFERENCES sales_visits(id) ON DELETE SET NULL;
        END IF;
    END IF;
END $$;
