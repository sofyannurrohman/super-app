ALTER TABLE sales_returns
    DROP COLUMN IF EXISTS invoice_adjusted_at;

ALTER TABLE sales_return_items DROP COLUMN IF EXISTS sales_item_id;
