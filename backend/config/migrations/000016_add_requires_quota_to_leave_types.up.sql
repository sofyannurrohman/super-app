-- Add requires_quota column to leave_types
ALTER TABLE leave_types ADD COLUMN requires_quota BOOLEAN DEFAULT TRUE;

-- Seed new permit types for all companies
INSERT INTO leave_types (id, company_id, name, is_paid, requires_quota, default_quota)
SELECT 
    gen_random_uuid(), 
    c.id, 
    vals.name, 
    vals.is_paid, 
    vals.requires_quota,
    vals.default_quota
FROM companies c
CROSS JOIN (VALUES 
    ('Izin Sakit', true, false, 0),
    ('Izin Terkena Musibah', true, false, 0),
    ('Izin Lainnya', false, false, 0)
) AS vals(name, is_paid, requires_quota, default_quota)
ON CONFLICT DO NOTHING;
