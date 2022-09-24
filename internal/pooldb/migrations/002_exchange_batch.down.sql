ALTER TABLE balance_inputs DROP COLUMN pool_fees;

ALTER TABLE balance_inputs DROP CONSTRAINT fk_balance_inputs_batch_id;
ALTER TABLE balance_inputs DROP INDEX idx_balance_inputs_batch_id;
ALTER TABLE balance_inputs DROP COLUMN batch_id;

ALTER TABLE balance_inputs DROP CONSTRAINT fk_balance_inputs_balance_output_id;
ALTER TABLE balance_inputs DROP INDEX idx_balance_inputs_balance_output_id;
ALTER TABLE balance_inputs RENAME COLUMN balance_output_id TO output_balance_id;
ALTER TABLE balance_inputs ADD INDEX idx_balance_inputs_output_balance_id (output_balance_id);

ALTER TABLE balance_inputs DROP CONSTRAINT fk_balance_inputs_out_chain_id;
ALTER TABLE balance_inputs DROP INDEX idx_balance_inputs_out_chain_id;
ALTER TABLE balance_inputs DROP COLUMN out_chain_id;

DROP TABLE balance_outputs;
DROP TABLE payouts;
DROP TABLE exchange_withdrawals;
DROP TABLE exchange_trades;
DROP TABLE exchange_deposits;
DROP TABLE exchange_inputs;
DROP TABLE exchange_batches;
DROP TABLE utxos;