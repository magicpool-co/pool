ALTER TABLE balance_inputs ADD COLUMN mature bool NOT NULL DEFAULT FALSE AFTER pool_fees;
UPDATE balance_inputs SET mature = TRUE;

ALTER TABLE balance_outputs ADD COLUMN mature bool NOT NULL DEFAULT FALSE AFTER exchange_fees;
UPDATE balance_outputs SET mature = TRUE;