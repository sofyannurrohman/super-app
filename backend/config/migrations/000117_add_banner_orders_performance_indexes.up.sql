-- Add performance indexes on banner_orders table
CREATE INDEX IF NOT EXISTS idx_banner_orders_created_at ON banner_orders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_banner_orders_status ON banner_orders(status);
CREATE INDEX IF NOT EXISTS idx_banner_orders_employee_id ON banner_orders(employee_id);
CREATE INDEX IF NOT EXISTS idx_banner_orders_store_id ON banner_orders(store_id);
CREATE INDEX IF NOT EXISTS idx_banner_orders_editor_id ON banner_orders(editor_id);
CREATE INDEX IF NOT EXISTS idx_banner_orders_status_created_at ON banner_orders(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stores_assigned_employee_id ON stores(assigned_employee_id);
CREATE INDEX IF NOT EXISTS idx_stores_is_active ON stores(is_active);
