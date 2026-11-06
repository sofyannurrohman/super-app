DELETE FROM leave_types WHERE name IN ('Izin Sakit', 'Izin Terkena Musibah', 'Izin Lainnya');
ALTER TABLE leave_types DROP COLUMN IF EXISTS requires_quota;
