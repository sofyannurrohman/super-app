-- Migration: Create sales_tester_requests and delivery_batch_samples tables
-- Version: 000060

CREATE TABLE IF NOT EXISTS sales_tester_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    store_id UUID REFERENCES stores(id) ON DELETE SET NULL,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, APPROVED, IN_DELIVERY, DELIVERED, REJECTED
    delivery_batch_id UUID REFERENCES delivery_batches(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS delivery_batch_samples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_batch_id UUID NOT NULL REFERENCES delivery_batches(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    recipient_type VARCHAR(20) NOT NULL DEFAULT 'STORE', -- STORE, SALESMAN
    store_id UUID REFERENCES stores(id) ON DELETE SET NULL,
    salesman_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    sales_tester_request_id UUID REFERENCES sales_tester_requests(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uni_batch_sample UNIQUE (delivery_batch_id, product_id, recipient_type, store_id, salesman_id)
);

CREATE INDEX IF NOT EXISTS idx_sales_tester_requests_branch ON sales_tester_requests(branch_id);
CREATE INDEX IF NOT EXISTS idx_sales_tester_requests_employee ON sales_tester_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_sales_tester_requests_status ON sales_tester_requests(status);
CREATE INDEX IF NOT EXISTS idx_delivery_batch_samples_batch ON delivery_batch_samples(delivery_batch_id);
