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
	("ERGO", true, false, false),
	("ETH", false, false, true),
	("ETC", true, true, false),
	("FIRO", true, true, false),
	("FLUX", true, true, false),
	("RVN", true, true, false),
	("USDC", false, false, true);

CREATE TABLE nodes (
	id				int         	UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,

	chain_id		varchar(4)		NOT NULL,
	region			varchar(20)		NOT NULL,
	url				varchar(100)	NOT NULL,
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

INSERT INTO nodes (chain_id, region, url, mainnet, enabled, backup, active, synced, height) VALUES 
	("AE", "us-west-2", "node-0.ae.us-west-2.magicpool.co", true, true, true, false, false, 0),
	("AE", "us-west-2", "node-1.ae.us-west-2.magicpool.co", true, true, false, false, false, 0),
	("AE", "us-west-2", "node-0.ae-testnet.us-west-2.magicpool.co", false, true, true, false, false, 0),

	("CFX", "us-west-2", "node-0.cfx.us-west-2.magicpool.co", true, true, true, false, false, 0),
	("CFX", "us-west-2", "node-1.cfx.us-west-2.magicpool.co", true, true, false, false, false, 0),
	("CFX", "us-west-2", "node-0.cfx-testnet.us-west-2.magicpool.co", false, true, true, false, false, 0),

	("CTXC", "us-west-2", "node-0.ctxc.us-west-2.magicpool.co", true, true, true, false, false, 0),
	("CTXC", "us-west-2", "node-1.ctxc.us-west-2.magicpool.co", true, true, false, false, false, 0),
	("CTXC", "us-west-2", "node-0.ctxc-testnet.us-west-2.magicpool.co", false, false, true, false, false, 0),

	("ERGO", "us-west-2", "node-0.ergo.us-west-2.magicpool.co", true, true, true, false, false, 0),
	("ERGO", "us-west-2", "node-1.ergo.us-west-2.magicpool.co", true, true, false, false, false, 0),
	("ERGO", "us-west-2", "node-0.ergo-testnet.us-west-2.magicpool.co", false, false, true, false, false, 0),

	("ETC", "us-west-2", "node-0.etc.us-west-2.magicpool.co", true, true, true, false, false, 0),
	("ETC", "us-west-2", "node-1.etc.us-west-2.magicpool.co", true, true, false, false, false, 0),
	("ETC", "us-west-2", "node-0.etc-testnet.us-west-2.magicpool.co", false, true, true, false, false, 0),

	("FIRO", "us-west-2", "node-0.firo.us-west-2.magicpool.co", true, true, true, false, false, 0),
	("FIRO", "us-west-2", "node-1.firo.us-west-2.magicpool.co", true, true, false, false, false, 0),
	("FIRO", "us-west-2", "node-0.firo-testnet.us-west-2.magicpool.co", false, true, true, false, false, 0),

	("FLUX", "us-west-2", "node-0.flux.us-west-2.magicpool.co", true, true, true, false, false, 0),
	("FLUX", "us-west-2", "node-1.flux.us-west-2.magicpool.co", true, true, false, false, false, 0),
	("FLUX", "us-west-2", "node-0.flux-testnet.us-west-2.magicpool.co", false, true, true, false, false, 0),

	("RVN", "us-west-2", "node-0.rvn.us-west-2.magicpool.co", true, true, true, false, false, 0),
	("RVN", "us-west-2", "node-1.rvn.us-west-2.magicpool.co", true, true, false, false, false, 0),
	("RVN", "us-west-2", "node-0.rvn-testnet.us-west-2.magicpool.co", false, true, true, false, false, 0);

CREATE TABLE miners (
	id				int         	UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	address			varchar(100)	NOT NULL,

	active			bool			NOT NULL,
	last_login		datetime,
	last_share		datetime,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_miners_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),

	INDEX idx_miners_chain_id (chain_id),
	INDEX idx_miners_chain_id_address (chain_id, address)
);

CREATE TABLE workers (
	id				int         	UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	miner_id		int				UNSIGNED NOT NULL,
	name			varchar(32)		NOT NULL,

	active			bool			NOT NULL,
	last_login		datetime,
	last_share		datetime,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_workers_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_workers_miner_id (miner_id),
	INDEX idx_workers_miner_id_name (miner_id, name)
);

CREATE TABLE ip_addresses (
	id				bigint         	UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	miner_id		int				UNSIGNED NOT NULL,

	ip_address		varchar(40)		NOT NULL,
	active			bool			NOT NULL,
	last_share		datetime		NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_ip_addresses_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_ip_addresses_miner_id (miner_id),
	INDEX idx_ip_addresses_last_share (last_share),
	UNIQUE INDEX idx_uq_ip_addresses_miner_id_ip_address (miner_id, ip_address)
);

CREATE TABLE rounds (
	id				int				UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,
	worker_id		int				UNSIGNED,

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
	CONSTRAINT fk_rounds_worker_id
	FOREIGN KEY (worker_id)			REFERENCES	workers(id),

	INDEX idx_rounds_chain_id (chain_id),
	INDEX idx_rounds_miner_id (miner_id),
	INDEX idx_rounds_worker_id (worker_id),
	INDEX idx_rounds_chain_id_height (chain_id, height DESC)
);

CREATE TABLE shares (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	round_id		int         	UNSIGNED NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,
	worker_id		int				UNSIGNED,

	count			bigint			UNSIGNED NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_shares_round_id
	FOREIGN KEY (round_id)			REFERENCES	rounds(id),
	CONSTRAINT fk_shares_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),
	CONSTRAINT fk_shares_worker_id
	FOREIGN KEY (worker_id)			REFERENCES	workers(id),

	INDEX idx_shares_round_id (round_id),
	INDEX idx_shares_miner_id (miner_id),
	INDEX idx_shares_worker_id (worker_id)
);
