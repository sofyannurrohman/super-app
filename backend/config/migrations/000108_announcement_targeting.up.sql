-- Tambah kolom targeting dan popup ke tabel announcements
ALTER TABLE announcements
    ADD COLUMN IF NOT EXISTS target_type VARCHAR(50) DEFAULT 'broadcast',
    ADD COLUMN IF NOT EXISTS is_popup BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS popup_expires_at TIMESTAMPTZ;

-- Tabel untuk menyimpan target user/role pilihan admin
CREATE TABLE IF NOT EXISTS announcement_targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    target_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    target_role VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_announcement_targets_announcement_id
    ON announcement_targets(announcement_id);
CREATE INDEX IF NOT EXISTS idx_announcement_targets_user_id
    ON announcement_targets(target_user_id);
CREATE INDEX IF NOT EXISTS idx_announcement_targets_role
    ON announcement_targets(target_role);

-- Tabel tracking siapa yang sudah acknowledge popup
CREATE TABLE IF NOT EXISTS announcement_reads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    acknowledged_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(announcement_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_announcement_reads_user_id
    ON announcement_reads(user_id);
CREATE INDEX IF NOT EXISTS idx_announcement_reads_announcement_id
    ON announcement_reads(announcement_id);
