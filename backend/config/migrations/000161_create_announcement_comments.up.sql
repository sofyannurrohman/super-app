-- Migration 000161: Create announcement_comments table and add comments_count to announcements
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS comments_count INT DEFAULT 0;

CREATE TABLE IF NOT EXISTS announcement_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_announcement_comments_announcement_id ON announcement_comments(announcement_id);
CREATE INDEX IF NOT EXISTS idx_announcement_comments_user_id ON announcement_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_announcement_comments_announcement_created ON announcement_comments(announcement_id, created_at ASC);
