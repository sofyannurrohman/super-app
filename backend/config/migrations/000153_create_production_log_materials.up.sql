-- +migrate Up

CREATE TABLE IF NOT EXISTS production_log_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    production_log_id UUID NOT NULL REFERENCES production_logs(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    batch_no VARCHAR(50) NOT NULL DEFAULT 'DEFAULT',
    quantity INT NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_production_log_materials_production_log_id
    ON production_log_materials(production_log_id);
