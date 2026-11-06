CREATE TABLE IF NOT EXISTS mt_store_stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 0,
    min_stock_threshold INT NOT NULL DEFAULT 20,
    last_adjusted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_mt_store_stocks_store_product UNIQUE (store_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_mt_store_stocks_store ON mt_store_stocks(store_id);
CREATE INDEX IF NOT EXISTS idx_mt_store_stocks_product ON mt_store_stocks(product_id);
CREATE INDEX IF NOT EXISTS idx_mt_store_stocks_company ON mt_store_stocks(company_id);

CREATE TABLE IF NOT EXISTS mt_store_stock_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- INITIAL, DELIVERY_IN, SPG_SALE_OUT, ADJUSTMENT, RETURN
    quantity INT NOT NULL,
    quantity_before INT NOT NULL DEFAULT 0,
    quantity_after INT NOT NULL DEFAULT 0,
    reference_no VARCHAR(100),
    notes TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mt_store_stock_logs_store ON mt_store_stock_logs(store_id);

-- Add TOP and Billing status fields to mt_purchase_orders if not present
ALTER TABLE mt_purchase_orders
ADD COLUMN IF NOT EXISTS payment_term VARCHAR(50) DEFAULT 'TOP_30_DAYS',
ADD COLUMN IF NOT EXISTS due_date DATE,
ADD COLUMN IF NOT EXISTS billing_status VARCHAR(50) DEFAULT 'UNBILLED',
ADD COLUMN IF NOT EXISTS payment_proof_url VARCHAR(255),
ADD COLUMN IF NOT EXISTS collected_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS collected_by VARCHAR(255);
