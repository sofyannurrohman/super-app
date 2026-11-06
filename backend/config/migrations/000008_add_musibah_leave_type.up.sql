-- Seed Musibah Type for all companies
INSERT INTO leave_types (id, company_id, name, is_paid, default_quota)
SELECT 
    gen_random_uuid(), 
    c.id, 
    'Izin Musibah / Duka (Dibayar)', 
    true, 
    3
FROM companies c
ON CONFLICT DO NOTHING;

-- Seed balances for existing employees for the new type
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
WHERE lt.name = 'Izin Musibah / Duka (Dibayar)'
ON CONFLICT (employee_id, leave_type_id, year) DO NOTHING;
