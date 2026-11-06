DROP TABLE IF EXISTS announcement_reads;
DROP TABLE IF EXISTS announcement_targets;
ALTER TABLE announcements
    DROP COLUMN IF EXISTS target_type,
    DROP COLUMN IF EXISTS is_popup,
    DROP COLUMN IF EXISTS popup_expires_at;
