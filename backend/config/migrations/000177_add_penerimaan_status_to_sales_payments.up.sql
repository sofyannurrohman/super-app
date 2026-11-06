ALTER TABLE sales_payments
ADD COLUMN IF NOT EXISTS penerimaan_status VARCHAR(50) DEFAULT 'PENDING',
ADD COLUMN IF NOT EXISTS penerimaan_verified_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS penerimaan_verified_by_id UUID REFERENCES employees(id),
ADD COLUMN IF NOT EXISTS penerimaan_notes TEXT;

CREATE INDEX IF NOT EXISTS idx_sales_payments_penerimaan_status ON sales_payments(penerimaan_status);
