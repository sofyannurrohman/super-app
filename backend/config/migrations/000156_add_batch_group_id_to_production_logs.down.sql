-- +migrate Down

DROP INDEX IF EXISTS idx_production_logs_batch_group_id;

ALTER TABLE production_logs DROP COLUMN IF EXISTS batch_group_id;
