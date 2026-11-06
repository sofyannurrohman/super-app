-- Fix requires_quota for all Izin types to be FALSE
UPDATE leave_types 
SET requires_quota = FALSE, default_quota = 0 
WHERE name ILIKE '%Izin Sakit%' 
   OR name ILIKE '%Izin Terkena Musibah%' 
   OR name ILIKE '%Izin Lainnya%'
   OR name ILIKE '%Izin Menikah%'
   OR name ILIKE '%Izin Melahirkan%';
