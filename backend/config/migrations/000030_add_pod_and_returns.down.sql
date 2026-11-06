-- Rollback POD and Sales Returns
-- Version: 000030

DROP TABLE IF EXISTS sales_return_items;
DROP TABLE IF EXISTS sales_returns;

DROP INDEX IF EXISTS idx_branch_product_type;
ALTER TABLE warehouse_stocks DROP COLUMN IF EXISTS stock_type;
CREATE UNIQUE INDEX IF NOT EXISTS idx_branch_product ON warehouse_stocks(branch_id, product_id);

ALTER TABLE sales_orders 
    DROP COLUMN IF EXISTS pod_image_url,
    DROP COLUMN IF EXISTS received_at,
    DROP COLUMN IF EXISTS received_by;
