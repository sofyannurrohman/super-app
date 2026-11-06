ALTER TABLE employee_kpis ADD COLUMN on_time_count INT DEFAULT 0;
ALTER TABLE employee_kpis ADD COLUMN late_count INT DEFAULT 0;
ALTER TABLE employee_kpis ADD COLUMN alpha_count INT DEFAULT 0;
ALTER TABLE employee_kpis ADD COLUMN permit_count INT DEFAULT 0;
ALTER TABLE employee_kpis ADD COLUMN status VARCHAR(20) DEFAULT 'DRAFT';
