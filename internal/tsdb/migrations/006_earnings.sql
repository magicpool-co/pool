CREATE TABLE global_earnings (
	chain_id				varchar(5)		NOT NULL,

	value					double			NOT NULL,
	avg_value				double			NOT NULL,

	pending					bool			NOT NULL,
	count					bigint			UNSIGNED NOT NULL,
	period					tinyint(1)		UNSIGNED NOT NULL,
	start_time				datetime		NOT NULL,
	end_time				datetime		NOT NULL,

	PRIMARY KEY (end_time, chain_id, period)
);

CREATE TABLE miner_earnings (
	chain_id				varchar(5)		NOT NULL,
	miner_id				int				UNSIGNED NOT NULL,

	value					double			NOT NULL,
	avg_value				double			NOT NULL,

	pending					bool			NOT NULL,
	count					bigint			UNSIGNED NOT NULL,
	period					tinyint(1)		UNSIGNED NOT NULL,
	start_time				datetime		NOT NULL,
	end_time				datetime		NOT NULL,

	PRIMARY KEY (end_time, chain_id, period)
);
