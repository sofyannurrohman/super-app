ALTER TABLE sales_orders ADD COLUMN is_physical_nota BOOLEAN DEFAULT FALSE;
ALTER TABLE sales_orders ADD COLUMN physical_nota_image_url VARCHAR(255);
