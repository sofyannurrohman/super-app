DROP TABLE IF EXISTS announcement_likes;
ALTER TABLE announcements DROP COLUMN IF EXISTS media_type;
ALTER TABLE announcements DROP COLUMN IF EXISTS image_url;
ALTER TABLE announcements DROP COLUMN IF EXISTS media_urls;
ALTER TABLE announcements DROP COLUMN IF EXISTS external_url;
ALTER TABLE announcements DROP COLUMN IF EXISTS likes_count;
