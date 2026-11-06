-- Seed Absence Types for all companies
INSERT INTO leave_types (id, company_id, name, is_paid, default_quota)
SELECT 
    gen_random_uuid(), 
    c.id, 
    vals.name, 
    vals.is_paid, 
    vals.default_quota
FROM companies c
CROSS JOIN (VALUES 
    ('Sakit (Dibayar)', true, 12),
    ('Izin (Potong Gaji)', false, 0)
) AS vals(name, is_paid, default_quota)
ON CONFLICT DO NOTHING;

-- Seed balances for all existing employees for the new types
INSERT INTO leave_balances (id, employee_id, leave_type_id, year, balance_total, balance_used)
SELECT 
    gen_random_uuid(), 
    e.id, 
    lt.id, 
    EXTRACT(YEAR FROM NOW())::INT, 
    lt.default_quota, 
    0
FROM employees e
JOIN leave_types lt ON lt.company_id = e.company_id
WHERE lt.name IN ('Sakit (Dibayar)', 'Izin (Potong Gaji)')
ON CONFLICT (employee_id, leave_type_id, year) DO NOTHING;
