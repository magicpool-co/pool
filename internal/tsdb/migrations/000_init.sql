CREATE TABLE raw_blocks (
	id						bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	chain_id				varchar(5)		NOT NULL,

	height					bigint			UNSIGNED NOT NULL,
	value					double			NOT NULL,
	difficulty				double			NOT NULL,
	uncle_count				bigint			UNSIGNED NOT NULL,
	tx_count				bigint			UNSIGNED NOT NULL,

	timestamp				datetime(3)		NOT NULL
);

CREATE TABLE blocks (
	chain_id				varchar(5)		NOT NULL,

	value					double			NOT NULL,
	difficulty				double			NOT NULL,
	block_time				double			NOT NULL,
	hashrate				double			NOT NULL,
	uncle_rate				double			NOT NULL,
	profitability			double			NOT NULL,
	avg_profitability		double			NOT NULL,

	pending					bool			NOT NULL,
	count					bigint			UNSIGNED NOT NULL,
	uncle_count				bigint			UNSIGNED NOT NULL,
	tx_count				bigint			UNSIGNED NOT NULL,
	period					tinyint(1)		UNSIGNED NOT NULL,
	start_time				datetime		NOT NULL,
	end_time				datetime		NOT NULL,

	PRIMARY KEY (end_time, chain_id, period)
);

CREATE TABLE rounds (
	chain_id				varchar(5)		NOT NULL,

	value					double			NOT NULL,
	difficulty				double			NOT NULL,
	round_time				double			NOT NULL,
	accepted_shares			double			NOT NULL,
	rejected_shares			double			NOT NULL,
	invalid_shares 			double 			NOT NULL,
	hashrate				double			NOT NULL,
	uncle_rate				double			NOT NULL,
	luck					double			NOT NULL,
	avg_luck				double			NOT NULL,
	profitability			double			NOT NULL,
	avg_profitability		double			NOT NULL,

	pending					bool			NOT NULL,
	count					bigint			UNSIGNED NOT NULL,
	uncle_count				bigint			UNSIGNED NOT NULL,
	period					tinyint(1)		UNSIGNED NOT NULL,
	start_time				datetime		NOT NULL,
	end_time				datetime		NOT NULL,

	PRIMARY KEY (end_time, chain_id, period)
);


CREATE TABLE global_shares (
	chain_id				varchar(5)		NOT NULL,

	miners					int				UNSIGNED NOT NULL,
	workers					int				UNSIGNED NOT NULL,
	accepted_shares			bigint			UNSIGNED NOT NULL,
	rejected_shares			bigint			UNSIGNED NOT NULL,
	invalid_shares 			bigint 			UNSIGNED NOT NULL,
	hashrate				double			NOT NULL,
	avg_hashrate			double			NOT NULL,
	reported_hashrate		double			NOT NULL,

	pending					bool			NOT NULL,
	count					bigint			UNSIGNED NOT NULL,
	period					tinyint(1)		UNSIGNED NOT NULL,
	start_time				datetime		NOT NULL,
	end_time				datetime		NOT NULL,

	PRIMARY KEY (end_time, chain_id, period)
);

CREATE TABLE miner_shares (
	chain_id				varchar(5)		NOT NULL,
	miner_id				int				UNSIGNED NOT NULL,

	workers					int				UNSIGNED NOT NULL,
	accepted_shares			bigint			UNSIGNED NOT NULL,
	rejected_shares			bigint			UNSIGNED NOT NULL,
	invalid_shares 			bigint 			UNSIGNED NOT NULL,
	hashrate				double			NOT NULL,
	avg_hashrate			double			NOT NULL,
	reported_hashrate		double			NOT NULL,

	pending					bool			NOT NULL,
	count					bigint			UNSIGNED NOT NULL,
	period					tinyint(1)		UNSIGNED NOT NULL,
	start_time				datetime		NOT NULL,
	end_time				datetime		NOT NULL,

	PRIMARY KEY (end_time, miner_id, chain_id, period),
	INDEX idx_miner_shares_miner_id_chain_id_period (miner_id, chain_id, period)
);

CREATE TABLE worker_shares (
	chain_id				varchar(5)		NOT NULL,
	worker_id				int				UNSIGNED NOT NULL,

	accepted_shares			bigint			UNSIGNED NOT NULL,
	rejected_shares			bigint			UNSIGNED NOT NULL,
	invalid_shares 			bigint 			UNSIGNED NOT NULL,
	hashrate				double			NOT NULL,
	avg_hashrate			double			NOT NULL,
	reported_hashrate		double			NOT NULL,

	pending					bool			NOT NULL,
	count					bigint			UNSIGNED NOT NULL,
	period					tinyint(1)		UNSIGNED NOT NULL,
	start_time				datetime		NOT NULL,
	end_time				datetime		NOT NULL,

	PRIMARY KEY (end_time, worker_id, chain_id, period),
	INDEX idx_worker_shares_worker_id_chain_id_period (worker_id, chain_id, period)
);
