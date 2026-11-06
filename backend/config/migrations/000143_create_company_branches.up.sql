-- Relasi operasional PT dan cabang.
-- Tetap mempertahankan branches.company_id untuk kompatibilitas modul lama.
CREATE TABLE IF NOT EXISTS company_branches (
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, branch_id)
);

CREATE INDEX IF NOT EXISTS idx_company_branches_branch_active
    ON company_branches (branch_id, is_active);

CREATE INDEX IF NOT EXISTS idx_company_branches_company_active
    ON company_branches (company_id, is_active);

-- Backfill relasi lama yang masih disimpan pada branches.company_id.
INSERT INTO company_branches (company_id, branch_id)
SELECT company_id, id
FROM branches
WHERE company_id IS NOT NULL
ON CONFLICT (company_id, branch_id) DO NOTHING;
