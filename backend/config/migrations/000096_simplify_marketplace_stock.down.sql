ALTER TABLE marketplace_stocks 
  ADD COLUMN qty_unpacked INT DEFAULT 0,
  ADD COLUMN qty_packed INT DEFAULT 0;

ALTER TABLE marketplace_opnames
  ADD COLUMN location VARCHAR(20) DEFAULT '';
