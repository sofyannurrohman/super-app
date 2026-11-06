-- Add ON DELETE CASCADE to attendance_logs
ALTER TABLE attendance_logs DROP CONSTRAINT IF EXISTS attendance_logs_employee_id_fkey;
ALTER TABLE attendance_logs ADD CONSTRAINT attendance_logs_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to employee_shifts
ALTER TABLE employee_shifts DROP CONSTRAINT IF EXISTS employee_shifts_employee_id_fkey;
ALTER TABLE employee_shifts ADD CONSTRAINT employee_shifts_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to leave_balances
ALTER TABLE leave_balances DROP CONSTRAINT IF EXISTS leave_balances_employee_id_fkey;
ALTER TABLE leave_balances ADD CONSTRAINT leave_balances_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to leave_requests
ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS leave_requests_employee_id_fkey;
ALTER TABLE leave_requests ADD CONSTRAINT leave_requests_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

-- Add ON DELETE SET NULL to leave_requests approver
ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS leave_requests_approved_by_fkey;
ALTER TABLE leave_requests ADD CONSTRAINT leave_requests_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES employees(id) ON DELETE SET NULL;

-- Add ON DELETE CASCADE to employee_salary_settings
ALTER TABLE employee_salary_settings DROP CONSTRAINT IF EXISTS employee_salary_settings_employee_id_fkey;
ALTER TABLE employee_salary_settings ADD CONSTRAINT employee_salary_settings_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to payslips
ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_employee_id_fkey;
ALTER TABLE payslips ADD CONSTRAINT payslips_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to overtimes
ALTER TABLE overtimes DROP CONSTRAINT IF EXISTS overtimes_employee_id_fkey;
ALTER TABLE overtimes ADD CONSTRAINT overtimes_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

-- Add ON DELETE SET NULL to overtimes approver
ALTER TABLE overtimes DROP CONSTRAINT IF EXISTS overtimes_approved_by_fkey;
ALTER TABLE overtimes ADD CONSTRAINT overtimes_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES employees(id) ON DELETE SET NULL;

-- Add ON DELETE SET NULL to reimbursements approver
ALTER TABLE reimbursements DROP CONSTRAINT IF EXISTS reimbursements_approved_by_fkey;
ALTER TABLE reimbursements ADD CONSTRAINT reimbursements_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES employees(id) ON DELETE SET NULL;

-- Add ON DELETE SET NULL to employees manager_id
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_manager_id_fkey;
ALTER TABLE employees ADD CONSTRAINT employees_manager_id_fkey FOREIGN KEY (manager_id) REFERENCES employees(id) ON DELETE SET NULL;

-- Ensure Performance tables exist with correct cascading deletes
CREATE TABLE IF NOT EXISTS sales_kpis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    target_omzet DECIMAL(15, 2) NOT NULL,
    achieved_omzet DECIMAL(15, 2) DEFAULT 0,
    estimated_bonus DECIMAL(15, 2) DEFAULT 0,
    period_month INT NOT NULL,
    period_year INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS employee_kpis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    attendance_score DECIMAL(5, 2) DEFAULT 0,
    productivity_score DECIMAL(5, 2) DEFAULT 0,
    final_score DECIMAL(5, 2) DEFAULT 0,
    period_month INT NOT NULL,
    period_year INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS managerial_appraisals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    manager_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    review_notes TEXT,
    rating DECIMAL(3, 2) NOT NULL,
    review_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
