-- Ensure quantity is perfectly synchronized before dropping columns
UPDATE marketplace_stocks
SET quantity = COALESCE(qty_unpacked, 0) + COALESCE(qty_packed, 0);

ALTER TABLE marketplace_stocks 
  DROP COLUMN qty_unpacked,
  DROP COLUMN qty_packed;

ALTER TABLE marketplace_opnames
  DROP COLUMN location;
