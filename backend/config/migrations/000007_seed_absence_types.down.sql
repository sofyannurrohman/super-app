DELETE FROM leave_balances WHERE leave_type_id IN (SELECT id FROM leave_types WHERE name IN ('Sakit (Dibayar)', 'Izin (Potong Gaji)'));
DELETE FROM leave_types WHERE name IN ('Sakit (Dibayar)', 'Izin (Potong Gaji)');
