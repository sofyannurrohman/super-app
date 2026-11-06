DROP TABLE IF EXISTS branding_campaign_videos;
DO $$ BEGIN
    IF to_regclass('public.app_menus') IS NOT NULL THEN
        IF to_regclass('public.access_role_menus') IS NOT NULL THEN
            DELETE FROM access_role_menus WHERE app_menu_id IN (SELECT id FROM app_menus WHERE path = '/sales/video-kampanye');
        END IF;
        DELETE FROM app_menus WHERE path = '/sales/video-kampanye';
    END IF;
END $$;
