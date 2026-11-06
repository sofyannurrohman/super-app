-- Migration: Create sales transaction adjustments and add verification audit columns
-- Version: 000050

ALTER TABLE sales_transactions 
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS verified_by_id UUID REFERENCES employees(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS sales_transaction_adjustment_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_transaction_id UUID NOT NULL REFERENCES sales_transactions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    original_quantity INT NOT NULL,
    adjusted_quantity INT NOT NULL,
    difference INT NOT NULL,
    notes TEXT,
    admin_nota_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_st_adj_transaction ON sales_transaction_adjustment_logs(sales_transaction_id);
