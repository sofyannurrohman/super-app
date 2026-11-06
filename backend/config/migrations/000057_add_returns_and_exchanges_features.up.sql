-- Migration: Add return type and product exchange features
-- Version: 000057

-- 1. Add return_type to sales_returns
ALTER TABLE sales_returns ADD COLUMN IF NOT EXISTS return_type VARCHAR(30) DEFAULT 'POTONG_NOTA';

-- 2. Create product_exchanges table for standalone warranty claims/exchanges
CREATE TABLE IF NOT EXISTS product_exchanges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange_no VARCHAR(50) NOT NULL UNIQUE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, APPROVED, COMPLETED, REJECTED
    delivery_type VARCHAR(20) NOT NULL DEFAULT 'DELIVERY', -- DIRECT (tukar langsung di tempat), DELIVERY (kirim via SJ)
    notes TEXT,
    approved_at TIMESTAMPTZ,
    approved_by_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Create product_exchange_items table
CREATE TABLE IF NOT EXISTS product_exchange_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_exchange_id UUID NOT NULL REFERENCES product_exchanges(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    reason VARCHAR(255),
    condition VARCHAR(20) NOT NULL DEFAULT 'DAMAGED'
);

-- 4. Create sales_subsequent_items table to track items to deliver in future SJ
CREATE TABLE IF NOT EXISTS sales_subsequent_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_order_id UUID REFERENCES sales_orders(id) ON DELETE CASCADE, -- Nullable if standalone exchange
    product_exchange_id UUID REFERENCES product_exchanges(id) ON DELETE CASCADE, -- Nullable if from sales return
    sales_return_id UUID REFERENCES sales_returns(id) ON DELETE SET NULL,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, IN_DELIVERY, SHIPPED, DELIVERED
    delivery_batch_id UUID REFERENCES delivery_batches(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Add sales_subsequent_item_id to delivery_items
ALTER TABLE delivery_items ADD COLUMN IF NOT EXISTS sales_subsequent_item_id UUID REFERENCES sales_subsequent_items(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sales_returns_type ON sales_returns(return_type);
CREATE INDEX IF NOT EXISTS idx_product_exchanges_store ON product_exchanges(store_id);
CREATE INDEX IF NOT EXISTS idx_product_exchanges_branch ON product_exchanges(branch_id);
CREATE INDEX IF NOT EXISTS idx_product_exchanges_status ON product_exchanges(status);
CREATE INDEX IF NOT EXISTS idx_sales_subsequent_items_so ON sales_subsequent_items(sales_order_id);
CREATE INDEX IF NOT EXISTS idx_sales_subsequent_items_exchange ON sales_subsequent_items(product_exchange_id);
CREATE INDEX IF NOT EXISTS idx_sales_subsequent_items_status ON sales_subsequent_items(status);
CREATE INDEX IF NOT EXISTS idx_delivery_items_subsequent ON delivery_items(sales_subsequent_item_id);
