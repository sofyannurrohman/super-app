-- Migration 000140 Down: Revert cascade foreign key on kas_advance_settlements

ALTER TABLE kas_advance_settlements DROP CONSTRAINT IF EXISTS fk_kas_transactions_settlement;
ALTER TABLE kas_advance_settlements DROP CONSTRAINT IF EXISTS fk_kas_advance_settlements_settled_by;

ALTER TABLE kas_advance_settlements
    ADD CONSTRAINT fk_kas_transactions_settlement
    FOREIGN KEY (kas_transaction_id)
    REFERENCES kas_transactions(id);
