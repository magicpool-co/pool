CREATE TABLE chains (
	id 				varchar(4) 		NOT NULL PRIMARY KEY,

	mineable		bool			NOT NULL,
	switchable		bool			NOT NULL,
	payable			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO chains (id, mineable, switchable, payable) VALUES 
	("AE", true, false, false),
	("BTC", false, false, true),
	("CFX", true, true, false),
	("CTXC", true, true, false),
	("ERG", true, false, false),
	("ETH", false, false, true),
	("ETC", true, true, false),
	("FIRO", true, true, false),
	("FLUX", true, true, false),
	("RVN", true, true, false),
	("KAS", true, true, true),
	("NEXA", true, true, true),
	("USDC", false, false, true);

CREATE TABLE nodes (
	url				varchar(50)		NOT NULL PRIMARY KEY,

	chain_id		varchar(4)		NOT NULL,
	region			varchar(20)		NOT NULL,
	version			varchar(50),

	mainnet			bool			NOT NULL,
	enabled			bool			NOT NULL,
	backup			bool			NOT NULL,
	active			bool			NOT NULL,
	synced			bool			NOT NULL,
	height			bigint,

	needs_backup	bool			NOT NULL DEFAULT FALSE,
	pending_backup	bool			NOT NULL DEFAULT FALSE,
	needs_update	bool			NOT NULL DEFAULT FALSE,
	pending_update	bool			NOT NULL DEFAULT FALSE,
	needs_resize	bool			NOT NULL DEFAULT FALSE,
	pending_resize	bool			NOT NULL DEFAULT FALSE,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	down_at			datetime,
	backup_at		datetime,

	CONSTRAINT fk_nodes_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),

	INDEX idx_nodes_chain_id (chain_id)
);

INSERT INTO nodes (url, chain_id, region, mainnet, enabled, backup, active, synced, height) VALUES 
	("node-0.cfx.eu-west-1.privatemagicpool.co", "CFX", "eu-west-1", true, true, true, false, false, 0),

	("node-0.ctxc.eu-west-1.privatemagicpool.co", "CTXC", "eu-west-1", true, true, true, false, false, 0),

	("node-0.erg.eu-west-1.privatemagicpool.co", "ERG", "eu-west-1", true, true, true, false, false, 0),
	("node-0.erg.eu-central-1.privatemagicpool.co", "ERG", "eu-central-1", true, true, false, false, false, 0),
	("node-0.erg.us-east-1.privatemagicpool.co", "ERG", "us-east-1", true, true, false, false, false, 0),
	("node-0.erg.us-west-2.privatemagicpool.co", "ERG", "us-west-2", true, true, false, false, false, 0),

	("node-0.etc.eu-west-1.privatemagicpool.co", "ETC", "eu-west-1", true, true, true, false, false, 0),
	("node-0.etc.eu-central-1.privatemagicpool.co", "ETC", "eu-central-1", true, true, false, false, false, 0),
	("node-0.etc.us-east-1.privatemagicpool.co", "ETC", "us-east-1", true, true, false, false, false, 0),
	("node-0.etc.us-west-2.privatemagicpool.co", "ETC", "us-west-2", true, true, false, false, false, 0),

	("node-0.firo.eu-west-1.privatemagicpool.co", "FIRO", "eu-west-1", true, true, true, false, false, 0),

	("node-0.flux.eu-west-1.privatemagicpool.co", "FLUX", "eu-west-1", true, true, true, false, false, 0),

	("node-0.kas.eu-west-1.privatemagicpool.co", "KAS", "eu-west-1", true, true, true, false, false, 0),
	("node-0.kas.eu-central-1.privatemagicpool.co", "KAS", "eu-central-1", true, true, false, false, false, 0),
	("node-0.kas.us-east-1.privatemagicpool.co", "KAS", "us-east-1", true, true, false, false, false, 0),
	("node-0.kas.us-west-2.privatemagicpool.co", "KAS", "us-west-2", true, true, false, false, false, 0),

	("node-0.nexa.eu-west-1.privatemagicpool.co", "NEXA", "eu-west-1", true, true, true, false, false, 0),
	("node-0.nexa.eu-central-1.privatemagicpool.co", "NEXA", "eu-central-1", true, true, false, false, false, 0),
	("node-0.nexa.us-east-1.privatemagicpool.co", "NEXA", "us-east-1", true, true, false, false, false, 0),
	("node-0.nexa.us-west-2.privatemagicpool.co", "NEXA", "us-west-2", true, true, false, false, false, 0),

	("node-0.rvn.eu-west-1.privatemagicpool.co", "RVN", "eu-west-1", true, true, true, false, false, 0);

CREATE TABLE miners (
	id				int         	UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	address			varchar(100)	NOT NULL,
	active			bool			NOT NULL,
	threshold 		decimal(25,0),

	recipient_fee_percent 	int UNSIGNED,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_miners_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),

	INDEX idx_miners_chain_id (chain_id),
	UNIQUE INDEX idx_uq_miners_chain_id_address (chain_id, address)
);

INSERT INTO 
	miners(chain_id, address, active, recipient_fee_percent) 
VALUES 
	("BTC", "bc1qf4aatnyyxldwhvnaa8fz5gsxq5ceu85lfgrpw6", true, 50),
	("BTC", "16CRhKimYsAy9wXZRXfDdockHcNx3s2h2D", true, 50);

CREATE TABLE workers (
	id				int         	UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	miner_id		int				UNSIGNED NOT NULL,
	name			varchar(32)		NOT NULL,
	active			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_workers_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_workers_miner_id (miner_id),
	UNIQUE INDEX idx_uq_workers_miner_id_name (miner_id, name)
);

CREATE TABLE ip_addresses (
	miner_id		int				UNSIGNED NOT NULL,
	worker_id		int				NOT NULL,
	chain_id		varchar(4)		NOT NULL,
	ip_address		varchar(40)		NOT NULL,

	active			bool			NOT NULL,
	expired			bool			NOT NULL,
	last_share		datetime		NOT NULL,
	rtt				float,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_ip_addresses_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_ip_addresses_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	PRIMARY KEY (miner_id, worker_id, chain_id, ip_address),
	INDEX idx_ip_addresses_chain_id (chain_id),
	INDEX idx_ip_addresses_miner_id (miner_id),
	INDEX idx_ip_addresses_last_share (last_share),
	INDEX idx_ip_addresses_active (active)
);

CREATE TABLE rounds (
	id				int				UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,

	height			bigint			UNSIGNED NOT NULL,
	uncle_height	bigint			UNSIGNED,
	epoch_height	bigint			UNSIGNED,

	hash			varchar(100)	NOT NULL,
	nonce			bigint			UNSIGNED,
	mix_digest		varchar(100),
	coinbase_txid	varchar(100),
	value			decimal(25,0),

	accepted_shares	bigint			UNSIGNED NOT NULL,
	rejected_shares	bigint			UNSIGNED NOT NULL,
	invalid_shares 	bigint 			UNSIGNED NOT NULL,

	difficulty		bigint			UNSIGNED NOT NULL,
	luck			float			NOT NULL,

	pending			bool			NOT NULL,
	uncle			bool			NOT NULL,
	orphan			bool			NOT NULL,
	mature			bool			NOT NULL,
	spent			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_rounds_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_rounds_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_rounds_chain_id (chain_id),
	INDEX idx_rounds_miner_id (miner_id),
	INDEX idx_rounds_chain_id_height (chain_id, height DESC),
	INDEX idx_rounds_created_at (created_at DESC),
	INDEX idx_rounds_miner_id_created_at (miner_id, created_at DESC)
);

CREATE TABLE shares (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	round_id		int         	UNSIGNED NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,

	count			bigint			UNSIGNED NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_shares_round_id
	FOREIGN KEY (round_id)			REFERENCES	rounds(id),
	CONSTRAINT fk_shares_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_shares_round_id (round_id),
	INDEX idx_shares_miner_id (miner_id)
);

CREATE TABLE transactions (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	type 			tinyint(1) 		UNSIGNED NOT NULL,

	txid			varchar(100)	NOT NULL,
	tx_hex			mediumtext		NOT NULL,
	height			bigint,
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

CREATE TABLE utxos (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	transaction_id 	bigint 			UNSIGNED,

	value			decimal(25,0)	NOT NULL,
	txid			varchar(100)	NOT NULL,
	idx				int				UNSIGNED NOT NULL,
	active 			boolean 		NOT NULL,
	spent			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_utxos_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_utxos_transaction_id
	FOREIGN KEY (transaction_id)	REFERENCES	transactions(id),

	INDEX idx_utxos_chain_id (chain_id),
	INDEX idx_utxos_transaction_id (transaction_id)
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
	transaction_id 	bigint			UNSIGNED,

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
	CONSTRAINT fk_exchange_deposits_transaction_id 
	FOREIGN KEY (transaction_id)		REFERENCES transactions(id),

	INDEX idx_exchange_deposits_batch_id (batch_id),
	INDEX idx_exchange_deposits_chain_id (chain_id),
	INDEX idx_exchange_deposits_network_id (network_id),
	INDEX idx_exchange_deposits_transaction_id (transaction_id)
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
	transaction_id 	bigint 			UNSIGNED,

	txid			varchar(100)	NOT NULL,
	height			bigint			UNSIGNED,

	value			decimal(25,0)	NOT NULL,
	fee_balance		decimal(25,0) 	NOT NULL,
	pool_fees		decimal(25,0)	NOT NULL,
	exchange_fees	decimal(25,0)	NOT NULL,
	tx_fees			decimal(25,0),
	pending 		bool	 		NOT NULL,
	confirmed		bool			NOT NULL,
	failed			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_payouts_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_payouts_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),
	CONSTRAINT fk_payouts_transaction_id
	FOREIGN KEY (transaction_id) 	REFERENCES transactions(id),

	INDEX idx_payouts_chain_id (chain_id),
	INDEX idx_payouts_miner_id (miner_id),
	INDEX idx_payouts_transaction_id (transaction_id)
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
	spent			boolean			NOT NULL,

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

CREATE TABLE balance_inputs (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	round_id		int         	UNSIGNED NOT NULL,
	chain_id		varchar(4)		NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,

	out_chain_id		varchar(4)	NOT NULL,
	balance_output_id	bigint 		UNSIGNED,
	batch_id			bigint 		UNSIGNED,

	value			decimal(25,0)	NOT NULL,
	pool_fees		decimal(25,0)	NOT NULL,
	pending			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_balance_inputs_round_id
	FOREIGN KEY (round_id)			REFERENCES	rounds(id),
	CONSTRAINT fk_balance_inputs_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_balance_inputs_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),
	CONSTRAINT fk_balance_inputs_out_chain_id
	FOREIGN KEY (out_chain_id)		REFERENCES	chains(id),
	CONSTRAINT fk_balance_inputs_balance_output_id
	FOREIGN KEY (balance_output_id)	REFERENCES	balance_outputs(id),
	CONSTRAINT fk_balance_inputs_batch_id
	FOREIGN KEY (batch_id) 			REFERENCES exchange_batches(id),

	INDEX idx_balance_inputs_round_id (round_id),
	INDEX idx_balance_inputs_chain_id (chain_id),
	INDEX idx_balance_inputs_miner_id (miner_id),
	INDEX idx_balance_inputs_out_chain_id (out_chain_id),
	INDEX idx_balance_inputs_balance_output_id (balance_output_id),
	INDEX idx_balance_inputs_batch_id (batch_id)
);