ALTER TABLE utxos DROP CONSTRAINT fk_utxos_transaction_id;
ALTER TABLE utxos DROP INDEX idx_utxos_transaction_id;
ALTER TABLE utxos DROP COLUMN transaction_id;

ALTER TABLE exchange_deposits ADD COLUMN spent boolean NOT NULL DEFAULT FALSE AFTER confirmed;
ALTER TABLE exchange_deposits DROP CONSTRAINT fk_exchange_deposits_transaction_id;
ALTER TABLE exchange_deposits DROP INDEX idx_exchange_deposits_transaction_id;
ALTER TABLE exchange_deposits DROP COLUMN transaction_id;

ALTER TABLE payouts DROP CONSTRAINT fk_payouts_transaction_id;
ALTER TABLE payouts DROP INDEX idx_payouts_transaction_id;
ALTER TABLE payouts DROP COLUMN transaction_id;

DROP TABLE transactions;
