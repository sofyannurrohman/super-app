CREATE TABLE IF NOT EXISTS branding_ai_design_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    banner_order_id UUID NOT NULL REFERENCES banner_orders(id) ON DELETE CASCADE,
    requested_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'queued_hermes',
    reference_banner_id UUID REFERENCES banner_orders(id) ON DELETE SET NULL,
    reference_image_url TEXT,
    prompt TEXT NOT NULL,
    svg_url TEXT,
    preview_url TEXT,
    cdr_url TEXT,
    cdr_file_name VARCHAR(500),
    error_message TEXT,
    claimed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_branding_ai_jobs_banner_created
    ON branding_ai_design_jobs (banner_order_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_branding_ai_jobs_status_created
    ON branding_ai_design_jobs (status, created_at ASC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_branding_ai_jobs_one_active_per_banner
    ON branding_ai_design_jobs (banner_order_id)
    WHERE status IN ('queued_hermes', 'hermes_processing', 'awaiting_corel', 'corel_processing');

