ALTER TABLE miners CHANGE COLUMN active active bool NOT NULL AFTER address;
ALTER TABLE miners DROP COLUMN last_login;
ALTER TABLE miners DROP COLUMN last_share;

ALTER TABLE workers DROP COLUMN last_login;
ALTER TABLE workers DROP COLUMN last_share;

DROP table ip_addresses;
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

ALTER TABLE rounds DROP CONSTRAINT fk_rounds_worker_id;
ALTER TABLE rounds DROP INDEX idx_rounds_worker_id;
ALTER TABLE rounds DROP COLUMN worker_id;
ALTER TABLE shares DROP CONSTRAINT fk_shares_worker_id;
ALTER TABLE shares DROP INDEX idx_shares_worker_id;
ALTER TABLE shares DROP COLUMN worker_id;