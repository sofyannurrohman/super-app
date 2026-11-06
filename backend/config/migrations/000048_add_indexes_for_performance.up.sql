CREATE INDEX IF NOT EXISTS idx_attendance_logs_employee_clock_in ON attendance_logs(employee_id, clock_in_time DESC);
CREATE INDEX IF NOT EXISTS idx_sales_visits_employee_date ON sales_visits(employee_id, check_in_time DESC);
CREATE INDEX IF NOT EXISTS idx_stores_assigned_employee ON stores(assigned_employee_id) WHERE assigned_employee_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sales_transactions_employee_date ON sales_transactions(employee_id, transaction_date DESC);
