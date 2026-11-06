ALTER TABLE attendance_logs 
ADD COLUMN IF NOT EXISTS is_out_of_town BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS destination VARCHAR(255) DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_attendance_logs_out_of_town ON attendance_logs(is_out_of_town);
