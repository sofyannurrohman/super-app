-- Revert ON DELETE behavior to restricted (default)
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_branch_id_fkey;
ALTER TABLE employees ADD CONSTRAINT employees_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches(id);

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_department_id_fkey;
ALTER TABLE employees ADD CONSTRAINT employees_department_id_fkey FOREIGN KEY (department_id) REFERENCES departments(id);

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_job_position_id_fkey;
ALTER TABLE employees ADD CONSTRAINT employees_job_position_id_fkey FOREIGN KEY (job_position_id) REFERENCES job_positions(id);

ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_parent_id_fkey;
ALTER TABLE departments ADD CONSTRAINT departments_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES departments(id);

ALTER TABLE branches DROP CONSTRAINT IF EXISTS branches_company_id_fkey;
ALTER TABLE branches ADD CONSTRAINT branches_company_id_fkey FOREIGN KEY (company_id) REFERENCES companies(id);

ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_company_id_fkey;
ALTER TABLE departments ADD CONSTRAINT departments_company_id_fkey FOREIGN KEY (company_id) REFERENCES companies(id);

ALTER TABLE job_positions DROP CONSTRAINT IF EXISTS job_positions_company_id_fkey;
ALTER TABLE job_positions ADD CONSTRAINT job_positions_company_id_fkey FOREIGN KEY (company_id) REFERENCES companies(id);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_company_id_fkey;
ALTER TABLE users ADD CONSTRAINT users_company_id_fkey FOREIGN KEY (company_id) REFERENCES companies(id);

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_company_id_fkey;
ALTER TABLE employees ADD CONSTRAINT employees_company_id_fkey FOREIGN KEY (company_id) REFERENCES companies(id);
