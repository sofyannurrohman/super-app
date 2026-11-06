ALTER TABLE employee_requests
ADD COLUMN IF NOT EXISTS requester_job_title VARCHAR(255),
ADD COLUMN IF NOT EXISTS hr_approver_id UUID REFERENCES employees(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS hr_approval_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS hr_approval_notes TEXT,
ADD COLUMN IF NOT EXISTS hr_status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
ADD COLUMN IF NOT EXISTS director_approver_id UUID REFERENCES employees(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS director_approval_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS director_approval_notes TEXT,
ADD COLUMN IF NOT EXISTS director_status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
ADD COLUMN IF NOT EXISTS rejected_by_role VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_employee_requests_hr_status ON employee_requests(hr_status);
CREATE INDEX IF NOT EXISTS idx_employee_requests_dir_status ON employee_requests(director_status);
