-- +migrate Down

DROP TABLE IF EXISTS production_log_operators;
ALTER TABLE production_logs DROP COLUMN IF EXISTS operator_name;
