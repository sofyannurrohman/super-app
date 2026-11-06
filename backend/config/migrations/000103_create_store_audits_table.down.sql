DROP TABLE IF EXISTS store_audits;

ALTER TABLE stores DROP COLUMN verification_status;

DROP TYPE IF EXISTS store_verification_status;
