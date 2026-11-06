-- Migration 000160: Add media and likes columns to announcements, and create announcement_likes table
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS media_type VARCHAR(50) DEFAULT 'none';
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS media_urls TEXT[];
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS external_url TEXT;
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS likes_count INT DEFAULT 0;

CREATE TABLE IF NOT EXISTS announcement_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uq_announcement_user_like UNIQUE (announcement_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_announcement_likes_announcement_id ON announcement_likes(announcement_id);
CREATE INDEX IF NOT EXISTS idx_announcement_likes_user_id ON announcement_likes(user_id);
