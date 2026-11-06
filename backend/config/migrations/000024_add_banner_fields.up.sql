ALTER TABLE banner_orders 
ADD COLUMN campaign_name VARCHAR(255),
ADD COLUMN location VARCHAR(255),
ADD COLUMN banner_type VARCHAR(100),
ADD COLUMN documentation_image_url TEXT;
