ALTER TABLE transactions ADD COLUMN type tinyint(1) UNSIGNED NOT NULL DEFAULT 0 AFTER chain_id;
ALTER TABLE payouts ADD COLUMN pending boolean NOT NULL DEFAULT FALSE AFTER tx_fees;