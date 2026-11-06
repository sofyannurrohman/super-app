CREATE SEQUENCE IF NOT EXISTS memo_requests_seq START WITH 300;

CREATE TABLE memo_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    destination VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    requester_name VARCHAR(255) NOT NULL,
    requester_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    approver_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    memo_number VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
