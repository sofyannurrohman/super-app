CREATE SEQUENCE IF NOT EXISTS employee_requests_seq START WITH 100;

CREATE TABLE IF NOT EXISTS employee_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_number VARCHAR(100) NOT NULL UNIQUE,
    requester_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    requester_name VARCHAR(255) NOT NULL,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    job_position_id UUID REFERENCES job_positions(id) ON DELETE SET NULL,
    custom_position_title VARCHAR(255),
    employment_type VARCHAR(50) NOT NULL DEFAULT 'FULL_TIME',
    request_reason VARCHAR(100) NOT NULL DEFAULT 'NEW_HEADCOUNT',
    replacement_for_name VARCHAR(255),
    headcount_needed INT NOT NULL DEFAULT 1,
    headcount_fulfilled INT NOT NULL DEFAULT 0,
    min_education VARCHAR(100),
    gender_preference VARCHAR(20) DEFAULT 'ANY',
    min_age INT DEFAULT 18,
    max_age INT DEFAULT 45,
    min_experience_years INT DEFAULT 0,
    required_skills TEXT,
    job_description TEXT,
    notes TEXT,
    target_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    approver_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    approval_date TIMESTAMP WITH TIME ZONE,
    approval_notes TEXT,
    reject_reason TEXT,
    recruitment_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_employee_requests_status ON employee_requests(status);
CREATE INDEX IF NOT EXISTS idx_employee_requests_dept ON employee_requests(department_id);
CREATE INDEX IF NOT EXISTS idx_employee_requests_branch ON employee_requests(branch_id);
CREATE INDEX IF NOT EXISTS idx_employee_requests_target_date ON employee_requests(target_date);
