-- 000180_create_factory_qc_samples.up.sql

CREATE TABLE IF NOT EXISTS factory_qc_samples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    factory_id UUID NOT NULL REFERENCES factories(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    production_log_id UUID REFERENCES production_logs(id) ON DELETE SET NULL,
    batch_no VARCHAR(50) NOT NULL,
    expiry_date TIMESTAMPTZ,
    production_date TIMESTAMPTZ,
    quantity INT NOT NULL DEFAULT 1,
    sample_type VARCHAR(50) NOT NULL DEFAULT 'LAB_QC',
    purpose TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    requested_by_id UUID NOT NULL,
    requested_by_name VARCHAR(100) NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_by_id UUID,
    approved_by_name VARCHAR(100),
    approved_at TIMESTAMPTZ,
    rejection_reason TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_factory_id ON factory_qc_samples(factory_id);
CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_product_id ON factory_qc_samples(product_id);
CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_production_log_id ON factory_qc_samples(production_log_id);
CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_status ON factory_qc_samples(status);
CREATE INDEX IF NOT EXISTS idx_factory_qc_samples_created_at ON factory_qc_samples(created_at DESC);
