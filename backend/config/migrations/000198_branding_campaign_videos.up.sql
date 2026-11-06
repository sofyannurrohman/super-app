CREATE TABLE IF NOT EXISTS branding_campaign_videos (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title varchar(200) NOT NULL,
    description text NOT NULL DEFAULT '',
    video_url text NOT NULL,
    product_name varchar(200) NOT NULL,
    purchase_url text NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_branding_campaign_videos_created_at ON branding_campaign_videos (created_at DESC);

DO $$
DECLARE
    branding_module_id uuid;
    campaign_menu_id uuid;
BEGIN
    IF to_regclass('public.app_modules') IS NULL OR to_regclass('public.app_menus') IS NULL OR to_regclass('public.access_role_menus') IS NULL THEN
        RETURN;
    END IF;
    SELECT id INTO branding_module_id FROM app_modules WHERE name = 'Branding' LIMIT 1;
    IF branding_module_id IS NULL THEN
        RETURN;
    END IF;
    SELECT id INTO campaign_menu_id FROM app_menus WHERE path = '/sales/video-kampanye' AND module_id = branding_module_id LIMIT 1;
    IF campaign_menu_id IS NULL THEN
        campaign_menu_id := gen_random_uuid();
        INSERT INTO app_menus (id, module_id, name, path, level, created_at)
        VALUES (campaign_menu_id, branding_module_id, 'Video Kampanye', '/sales/video-kampanye', 1, now());
    END IF;
    INSERT INTO access_role_menus (access_role_id, app_menu_id)
    SELECT DISTINCT arm.access_role_id, campaign_menu_id
    FROM access_role_menus arm
    JOIN app_menus old_menu ON old_menu.id = arm.app_menu_id
    WHERE old_menu.path IN ('/sales/operator', '/sales/editor', '/sales/spanduk-validator', '/sales/spanduk-vendor')
    ON CONFLICT DO NOTHING;
END $$;
