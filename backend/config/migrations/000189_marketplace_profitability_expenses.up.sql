CREATE TABLE marketplace_profitability_expenses (
    id BIGSERIAL PRIMARY KEY,
    branch_id UUID NOT NULL,
    marketplace VARCHAR(50) NOT NULL CHECK (marketplace IN ('tiktok', 'shopee')),
    expense_date DATE NOT NULL,
    category VARCHAR(30) NOT NULL CHECK (category IN ('ads', 'fulfillment', 'return', 'operational')),
    amount DECIMAL(15,2) NOT NULL CHECK (amount >= 0),
    attributed_revenue DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (attributed_revenue >= 0),
    seller_sku VARCHAR(100) NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_marketplace_profit_expenses_period ON marketplace_profitability_expenses (expense_date, branch_id, marketplace);
