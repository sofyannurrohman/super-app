CREATE TABLE epkwt_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    document_number VARCHAR(100),
    
    -- Personal Data
    identity_number VARCHAR(50) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    phone_number VARCHAR(20),
    gender VARCHAR(10),
    birth_date DATE,
    birth_place VARCHAR(100),
    address_ktp TEXT,
    address_residential TEXT,
    marital_status VARCHAR(20),
    
    -- Job Data (filled by HR)
    job_position_id UUID REFERENCES job_positions(id) ON DELETE SET NULL,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    salary DECIMAL(15,2) DEFAULT 0,
    start_date DATE,
    end_date DATE,
    agreement_date DATE,
    
    -- Signatures & status
    employee_signature_url TEXT,
    hr_signature_url TEXT,
    status VARCHAR(50) DEFAULT 'PENDING_HR', -- 'PENDING_HR', 'APPROVED_DRAFT', 'ACTIVE', 'REJECTED'
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
