CREATE TABLE prices (
	chain_id				varchar(5)		NOT NULL,

	price_usd				double			NOT NULL,
	price_btc				double			NOT NULL,
	price_eth				double			NOT NULL,

	timestamp				datetime		NOT NULL,

	PRIMARY KEY (chain_id, timestamp),
	INDEX idx_prices_chain_id (chain_id),
	INDEX idx_prices_timestamp (timestamp)
);