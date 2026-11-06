-- +migrate Down

DROP TABLE IF EXISTS supplier_stocks;
DROP TABLE IF EXISTS supplier_receipt_items;
DROP TABLE IF EXISTS supplier_receipts;
DROP TABLE IF EXISTS suppliers;
