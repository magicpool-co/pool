CREATE TABLE transactions (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,

	txid			varchar(100)	NOT NULL,
	tx_hex			text			NOT NULL,
	value			decimal(25,0)	NOT NULL,
	fee				decimal(25,0)	NOT NULL,
	fee_balance		decimal(25,0),
	remainder		decimal(25,0)	NOT NULL,
	remainder_idx	int				NOT NULL,
	spent			bool			NOT NULL,
	confirmed		bool			NOT NULL,
	failed			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_transactions_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),

	INDEX idx_transactions_chain_id (chain_id)
);

ALTER TABLE utxos ADD COLUMN transaction_id bigint UNSIGNED AFTER chain_id;
ALTER TABLE utxos ADD INDEX idx_utxos_transaction_id (transaction_id);
ALTER TABLE utxos ADD CONSTRAINT fk_utxos_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id);

ALTER TABLE exchange_deposits DROP COLUMN spent;
ALTER TABLE exchange_deposits ADD COLUMN transaction_id bigint UNSIGNED AFTER network_id;
ALTER TABLE exchange_deposits ADD INDEX idx_exchange_deposits_transaction_id (transaction_id);
ALTER TABLE exchange_deposits ADD CONSTRAINT fk_exchange_deposits_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id);

ALTER TABLE payouts ADD COLUMN transaction_id bigint UNSIGNED AFTER address;
ALTER TABLE payouts ADD INDEX idx_payouts_transaction_id (transaction_id);
ALTER TABLE payouts ADD CONSTRAINT fk_payouts_transaction_id FOREIGN KEY (transaction_id) REFERENCES transactions(id);
