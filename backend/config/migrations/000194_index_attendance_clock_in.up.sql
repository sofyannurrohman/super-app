-- The all-branch monthly filter and DESC listing both use clock_in_time.
-- CONCURRENTLY keeps attendance writes available while the index is built.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_attendance_logs_clock_in_time_desc
    ON attendance_logs (clock_in_time DESC);
