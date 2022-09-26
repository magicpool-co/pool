CREATE TABLE utxos (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,

	value			decimal(25,0)	NOT NULL,
	txid			varchar(100)	NOT NULL,
	idx				int				UNSIGNED NOT NULL,
	spent			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_utxos_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),

	INDEX idx_utxos_chain_id (chain_id)
);

CREATE TABLE exchange_batches (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	exchange_id		tinyint(1)		UNSIGNED NOT NULL,
	status			tinyint(1)		UNSIGNED NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	completed_at	datetime
);

CREATE TABLE exchange_inputs (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	batch_id		bigint			UNSIGNED NOT NULL,
	in_chain_id		varchar(4)		NOT NULL,
	out_chain_id	varchar(4)		NOT NULL,

	value			decimal(25,0)	NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_exchange_inputs_batch_id
	FOREIGN KEY (batch_id)				REFERENCES	exchange_batches(id),
	CONSTRAINT fk_exchange_inputs_in_chain_id
	FOREIGN KEY (in_chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_exchange_inputs_out_chain_id
	FOREIGN KEY (out_chain_id)			REFERENCES	chains(id),

	INDEX idx_exchange_inputs_batch_id (batch_id),
	INDEX idx_exchange_inputs_in_chain_id (in_chain_id),
	INDEX idx_exchange_inputs_out_chain_id (out_chain_id)
);

CREATE TABLE exchange_deposits (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	batch_id		bigint			UNSIGNED NOT NULL,
	chain_id		varchar(4)		NOT NULL,
	network_id		varchar(4)		NOT NULL,

	deposit_txid		varchar(100)	NOT NULL,
	exchange_txid		varchar(100),
	exchange_deposit_id	varchar(100),

	value			decimal(25,0)	NOT NULL,
	fees			decimal(25,0),
	registered		bool			NOT NULL,
	confirmed		bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_exchange_deposits_batch_id
	FOREIGN KEY (batch_id)				REFERENCES	exchange_batches(id),
	CONSTRAINT fk_exchange_deposits_chain_id
	FOREIGN KEY (chain_id)				REFERENCES	chains(id),
	CONSTRAINT fk_exchange_deposits_network_id
	FOREIGN KEY (network_id)			REFERENCES	chains(id),

	INDEX idx_exchange_deposits_batch_id (batch_id),
	INDEX idx_exchange_deposits_chain_id (chain_id),
	INDEX idx_exchange_deposits_network_id (network_id)
);

CREATE TABLE exchange_trades (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	batch_id		bigint			UNSIGNED NOT NULL,
	path_id			tinyint(1)		UNSIGNED NOT NULL,
	stage_id		tinyint(1)		UNSIGNED NOT NULL,

	exchange_trade_id	varchar(100),

	initial_chain_id	varchar(4)		NOT NULL,
	from_chain_id		varchar(4)		NOT NULL,
	to_chain_id			varchar(4)		NOT NULL,
	market				varchar(12)		NOT NULL,
	direction			tinyint(1)		NOT NULL,

	value					decimal(25,0),
	proceeds				decimal(25,0),
	trade_fees				decimal(25,0),
	cumulative_deposit_fees	decimal(25,0),
	cumulative_trade_fees	decimal(25,0),

	order_price				double,
	fill_price				double,
	cumulative_fill_price	double,
	slippage				double,
	initiated				bool		NOT NULL,
	confirmed				bool		NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_exchange_trades_batch_id
	FOREIGN KEY (batch_id)				REFERENCES	exchange_batches(id),
	CONSTRAINT fk_exchange_trades_from_chain_id
	FOREIGN KEY (from_chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_exchange_trades_to_chain_id
	FOREIGN KEY (to_chain_id)			REFERENCES	chains(id),

	INDEX idx_exchange_trades_batch_id (batch_id),
	INDEX idx_exchange_trades_from_chain_id (from_chain_id),
	INDEX idx_exchange_trades_to_chain_id (to_chain_id)
);

CREATE TABLE exchange_withdrawals (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	batch_id		bigint			UNSIGNED NOT NULL,
	chain_id		varchar(4)		NOT NULL,
	network_id		varchar(4)		NOT NULL,

	exchange_txid			varchar(100),
	exchange_withdrawal_id	varchar(100)	NOT NULL,

	value			decimal(25,0)	NOT NULL,
	deposit_fees	decimal(25,0),
	trade_fees		decimal(25,0),
	withdrawal_fees	decimal(25,0),
	cumulative_fees	decimal(25,0),
	confirmed		bool			NOT NULL,
	spent			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_exchange_withdrawals_batch_id
	FOREIGN KEY (batch_id)				REFERENCES	exchange_batches(id),
	CONSTRAINT fk_exchange_withdrawals_chain_id
	FOREIGN KEY (chain_id)				REFERENCES	chains(id),
	CONSTRAINT fk_exchange_withdrawals_network_id
	FOREIGN KEY (network_id)			REFERENCES	chains(id),

	INDEX idx_exchange_withdrawals_batch_id (batch_id),
	INDEX idx_exchange_withdrawals_chain_id (chain_id),
	INDEX idx_exchange_withdrawals_network_id (network_id)
);

CREATE TABLE payouts (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,
	address			varchar(100)	NOT NULL,

	txid			varchar(100)	NOT NULL,
	height			bigint			UNSIGNED,

	value			decimal(25,0)	NOT NULL,
	fee_balance		decimal(25,0) 	NOT NULL,
	pool_fees		decimal(25,0)	NOT NULL,
	exchange_fees	decimal(25,0)	NOT NULL,
	tx_fees			decimal(25,0),
	confirmed		bool			NOT NULL,
	failed			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_payouts_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_payouts_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_payouts_chain_id (chain_id),
	INDEX idx_payouts_miner_id (miner_id)
);

CREATE TABLE balance_outputs (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,

	in_batch_id		bigint			UNSIGNED,
	in_deposit_id	bigint			UNSIGNED,
	in_payout_id	bigint			UNSIGNED,
	out_payout_id	bigint			UNSIGNED,

	value			decimal(25,0)	NOT NULL,
	pool_fees		decimal(25,0)	NOT NULL,
	exchange_fees	decimal(25,0)	NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_balance_outputs_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_balance_outputs_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),
	CONSTRAINT fk_balance_inputs_in_deposit_id
	FOREIGN KEY (in_deposit_id)		REFERENCES	exchange_deposits(id),
	CONSTRAINT fk_balance_inputs_in_payout_id
	FOREIGN KEY (in_payout_id)		REFERENCES	payouts(id),
	CONSTRAINT fk_balance_inputs_out_payout_id
	FOREIGN KEY (out_payout_id)		REFERENCES	payouts(id),

	INDEX idx_balance_outputs_chain_id (chain_id),
	INDEX idx_balance_outputs_miner_id (miner_id),
	INDEX idx_balance_inputs_in_deposit_id (in_deposit_id),
	INDEX idx_balance_inputs_in_payout_id (in_payout_id),
	INDEX idx_balance_inputs_out_payout_id (out_payout_id)
);

ALTER TABLE balance_inputs ADD COLUMN out_chain_id varchar(4) NOT NULL AFTER miner_id;
ALTER TABLE balance_inputs ADD INDEX idx_balance_inputs_out_chain_id (out_chain_id);
ALTER TABLE balance_inputs ADD CONSTRAINT fk_balance_inputs_out_chain_id FOREIGN KEY (out_chain_id) REFERENCES chains(id);

ALTER TABLE balance_inputs DROP INDEX idx_balance_inputs_output_balance_id;
ALTER TABLE balance_inputs RENAME COLUMN output_balance_id TO balance_output_id;
ALTER TABLE balance_inputs ADD INDEX idx_balance_inputs_balance_output_id (balance_output_id);
ALTER TABLE balance_inputs ADD CONSTRAINT fk_balance_inputs_balance_output_id FOREIGN KEY (balance_output_id) REFERENCES balance_outputs(id);

ALTER TABLE balance_inputs ADD COLUMN batch_id bigint UNSIGNED AFTER balance_output_id;
ALTER TABLE balance_inputs ADD INDEX idx_balance_inputs_batch_id (batch_id);
ALTER TABLE balance_inputs ADD CONSTRAINT fk_balance_inputs_batch_id FOREIGN KEY (batch_id) REFERENCES exchange_batches(id);

ALTER TABLE balance_inputs ADD COLUMN pool_fees decimal(25,0) NOT NULL AFTER value;

