-- +migrate Up

ALTER TABLE production_logs ADD COLUMN IF NOT EXISTS batch_group_id UUID;

CREATE INDEX IF NOT EXISTS idx_production_logs_batch_group_id
    ON production_logs(batch_group_id);
