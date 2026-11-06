-- Revert ON DELETE CASCADE to restricted (default)
ALTER TABLE attendance_logs DROP CONSTRAINT IF EXISTS attendance_logs_employee_id_fkey;
ALTER TABLE attendance_logs ADD CONSTRAINT attendance_logs_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE employee_shifts DROP CONSTRAINT IF EXISTS employee_shifts_employee_id_fkey;
ALTER TABLE employee_shifts ADD CONSTRAINT employee_shifts_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE leave_balances DROP CONSTRAINT IF EXISTS leave_balances_employee_id_fkey;
ALTER TABLE leave_balances ADD CONSTRAINT leave_balances_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS leave_requests_employee_id_fkey;
ALTER TABLE leave_requests ADD CONSTRAINT leave_requests_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS leave_requests_approved_by_fkey;
ALTER TABLE leave_requests ADD CONSTRAINT leave_requests_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES employees(id);

ALTER TABLE employee_salary_settings DROP CONSTRAINT IF EXISTS employee_salary_settings_employee_id_fkey;
ALTER TABLE employee_salary_settings ADD CONSTRAINT employee_salary_settings_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE payslips DROP CONSTRAINT IF EXISTS payslips_employee_id_fkey;
ALTER TABLE payslips ADD CONSTRAINT payslips_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE overtimes DROP CONSTRAINT IF EXISTS overtimes_employee_id_fkey;
ALTER TABLE overtimes ADD CONSTRAINT overtimes_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE overtimes DROP CONSTRAINT IF EXISTS overtimes_approved_by_fkey;
ALTER TABLE overtimes ADD CONSTRAINT overtimes_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES employees(id);

ALTER TABLE reimbursements DROP CONSTRAINT IF EXISTS reimbursements_approved_by_fkey;
ALTER TABLE reimbursements ADD CONSTRAINT reimbursements_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES employees(id);

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_manager_id_fkey;
ALTER TABLE employees ADD CONSTRAINT employees_manager_id_fkey FOREIGN KEY (manager_id) REFERENCES employees(id);

-- Drop Performance tables
DROP TABLE IF EXISTS sales_kpis;
DROP TABLE IF EXISTS employee_kpis;
DROP TABLE IF EXISTS managerial_appraisals;
