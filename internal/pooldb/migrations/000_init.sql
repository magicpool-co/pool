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
	("node-0.cfx.us-west-2.privatemagicpool.co", "CFX", "us-west-2", true, true, false, false, false, 0),
	("node-1.cfx.us-west-2.privatemagicpool.co", "CFX", "us-west-2", true, true, false, false, false, 0),
	("node-0.cfx.eu-central-1.privatemagicpool.co", "CFX", "eu-central-1", true, true, false, false, false, 0),
	("node-1.cfx.eu-central-1.privatemagicpool.co", "CFX", "eu-central-1", true, true, false, false, false, 0),
	("node-0.cfx.ap-southeast-1.privatemagicpool.co", "CFX", "ap-southeast-1", true, true, false, false, false, 0),
	("node-1.cfx.ap-southeast-1.privatemagicpool.co", "CFX", "ap-southeast-1", true, true, false, false, false, 0),

	("node-0.ctxc.eu-west-1.privatemagicpool.co", "CTXC", "eu-west-1", true, true, true, false, false, 0),
	("node-0.ctxc.us-west-2.privatemagicpool.co", "CTXC", "us-west-2", true, true, false, false, false, 0),
	("node-1.ctxc.us-west-2.privatemagicpool.co", "CTXC", "us-west-2", true, true, false, false, false, 0),
	("node-0.ctxc.eu-central-1.privatemagicpool.co", "CTXC", "eu-central-1", true, true, false, false, false, 0),
	("node-1.ctxc.eu-central-1.privatemagicpool.co", "CTXC", "eu-central-1", true, true, false, false, false, 0),
	("node-0.ctxc.ap-southeast-1.privatemagicpool.co", "CTXC", "ap-southeast-1", true, true, false, false, false, 0),
	("node-1.ctxc.ap-southeast-1.privatemagicpool.co", "CTXC", "ap-southeast-1", true, true, false, false, false, 0),

	("node-0.ergo.eu-west-1.privatemagicpool.co", "ERGO", "eu-west-1", true, true, true, false, false, 0),
	("node-0.ergo.us-west-2.privatemagicpool.co", "ERGO", "us-west-2", true, true, false, false, false, 0),
	("node-1.ergo.us-west-2.privatemagicpool.co", "ERGO", "us-west-2", true, true, false, false, false, 0),
	("node-0.ergo.eu-central-1.privatemagicpool.co", "ERGO", "eu-central-1", true, true, false, false, false, 0),
	("node-1.ergo.eu-central-1.privatemagicpool.co", "ERGO", "eu-central-1", true, true, false, false, false, 0),
	("node-0.ergo.ap-southeast-1.privatemagicpool.co", "ERGO", "ap-southeast-1", true, true, false, false, false, 0),
	("node-1.ergo.ap-southeast-1.privatemagicpool.co", "ERGO", "ap-southeast-1", true, true, false, false, false, 0),

	("node-0.etc.eu-west-1.privatemagicpool.co", "ETC", "eu-west-1", true, true, true, false, false, 0),
	("node-0.etc.us-west-2.privatemagicpool.co", "ETC", "us-west-2", true, true, false, false, false, 0),
	("node-1.etc.us-west-2.privatemagicpool.co", "ETC", "us-west-2", true, true, false, false, false, 0),
	("node-0.etc.eu-central-1.privatemagicpool.co", "ETC", "eu-central-1", true, true, false, false, false, 0),
	("node-1.etc.eu-central-1.privatemagicpool.co", "ETC", "eu-central-1", true, true, false, false, false, 0),
	("node-0.etc.ap-southeast-1.privatemagicpool.co", "ETC", "ap-southeast-1", true, true, false, false, false, 0),
	("node-1.etc.ap-southeast-1.privatemagicpool.co", "ETC", "ap-southeast-1", true, true, false, false, false, 0),

	("node-0.firo.eu-west-1.privatemagicpool.co", "FIRO", "eu-west-1", true, true, true, false, false, 0),
	("node-0.firo.us-west-2.privatemagicpool.co", "FIRO", "us-west-2", true, true, false, false, false, 0),
	("node-1.firo.us-west-2.privatemagicpool.co", "FIRO", "us-west-2", true, true, false, false, false, 0),
	("node-0.firo.eu-central-1.privatemagicpool.co", "FIRO", "eu-central-1", true, true, false, false, false, 0),
	("node-1.firo.eu-central-1.privatemagicpool.co", "FIRO", "eu-central-1", true, true, false, false, false, 0),
	("node-0.firo.ap-southeast-1.privatemagicpool.co", "FIRO", "ap-southeast-1", true, true, false, false, false, 0),
	("node-1.firo.ap-southeast-1.privatemagicpool.co", "FIRO", "ap-southeast-1", true, true, false, false, false, 0),

	("node-0.flux.eu-west-1.privatemagicpool.co", "FLUX", "eu-west-1", true, true, true, false, false, 0),
	("node-0.flux.us-west-2.privatemagicpool.co", "FLUX", "us-west-2", true, true, false, false, false, 0),
	("node-1.flux.us-west-2.privatemagicpool.co", "FLUX", "us-west-2", true, true, false, false, false, 0),
	("node-0.flux.eu-central-1.privatemagicpool.co", "FLUX", "eu-central-1", true, true, false, false, false, 0),
	("node-1.flux.eu-central-1.privatemagicpool.co", "FLUX", "eu-central-1", true, true, false, false, false, 0),
	("node-0.flux.ap-southeast-1.privatemagicpool.co", "FLUX", "ap-southeast-1", true, true, false, false, false, 0),
	("node-1.flux.ap-southeast-1.privatemagicpool.co", "FLUX", "ap-southeast-1", true, true, false, false, false, 0),

	("node-0.rvn.eu-west-1.privatemagicpool.co", "RVN", "eu-west-1", true, true, true, false, false, 0),
	("node-0.rvn.us-west-2.privatemagicpool.co", "RVN", "us-west-2", true, true, false, false, false, 0),
	("node-1.rvn.us-west-2.privatemagicpool.co", "RVN", "us-west-2", true, true, false, false, false, 0),
	("node-0.rvn.eu-central-1.privatemagicpool.co", "RVN", "eu-central-1", true, true, false, false, false, 0),
	("node-1.rvn.eu-central-1.privatemagicpool.co", "RVN", "eu-central-1", true, true, false, false, false, 0),
	("node-0.rvn.ap-southeast-1.privatemagicpool.co", "RVN", "ap-southeast-1", true, true, false, false, false, 0),
	("node-1.rvn.ap-southeast-1.privatemagicpool.co", "RVN", "ap-southeast-1", true, true, false, false, false, 0);

CREATE TABLE miners (
	id				int         	UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id		varchar(4)		NOT NULL,
	address			varchar(100)	NOT NULL,
	active			bool			NOT NULL,

	recipient_fee_percent 	int UNSIGNED,

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

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_workers_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_workers_miner_id (miner_id),
	INDEX idx_workers_miner_id_name (miner_id, name)
);

CREATE TABLE ip_addresses (
	miner_id		int				UNSIGNED NOT NULL,
	worker_id		int				NOT NULL,
	chain_id		varchar(4)		NOT NULL,
	ip_address		varchar(40)		NOT NULL,

	active			bool			NOT NULL,
	last_share		datetime		NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_ip_addresses_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_ip_addresses_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	PRIMARY KEY (miner_id, worker_id, chain_id, ip_address),
	INDEX idx_ip_addresses_chain_id (chain_id),
	INDEX idx_ip_addresses_miner_id (miner_id),
	INDEX idx_ip_addresses_last_share (last_share)
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
	INDEX idx_rounds_chain_id_height (chain_id, height DESC)
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

CREATE TABLE balance_inputs (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	round_id		int         	UNSIGNED NOT NULL,
	chain_id		varchar(4)		NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,
	
	output_balance_id	bigint UNSIGNED,

	value			decimal(25,0)	NOT NULL,
	pending			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_balance_inputs_round_id
	FOREIGN KEY (round_id)			REFERENCES	rounds(id),
	CONSTRAINT fk_balance_inputs_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_balance_inputs_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_balance_inputs_round_id (round_id),
	INDEX idx_balance_inputs_chain_id (chain_id),
	INDEX idx_balance_inputs_miner_id (miner_id),
	INDEX idx_balance_inputs_output_balance_id (output_balance_id)
);