DROP INDEX IF EXISTS idx_attendance_logs_out_of_town;
ALTER TABLE attendance_logs 
DROP COLUMN IF EXISTS destination,
DROP COLUMN IF EXISTS is_out_of_town;
