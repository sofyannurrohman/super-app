CREATE TABLE IF NOT EXISTS employee_request_candidates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_request_id UUID NOT NULL REFERENCES employee_requests(id) ON DELETE CASCADE,
    candidate_name VARCHAR(255) NOT NULL,
    candidate_phone VARCHAR(50),
    candidate_email VARCHAR(255),
    join_date DATE NOT NULL,
    document_url TEXT,
    document_name VARCHAR(255),
    document_type VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_req_candidates_req_id ON employee_request_candidates(employee_request_id);
