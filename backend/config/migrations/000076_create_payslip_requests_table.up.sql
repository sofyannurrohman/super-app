CREATE TABLE IF NOT EXISTS payslip_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    period_month INT NOT NULL,
    period_year INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    nominal DECIMAL(15,2),
    file_url VARCHAR(255),
    reject_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
