ALTER TABLE miners CHANGE COLUMN active active bool NOT NULL AFTER recipient_fee_percent;
ALTER TABLE miners ADD COLUMN last_login datetime AFTER active;
ALTER TABLE miners ADD COLUMN last_share datetime AFTER last_login;

ALTER TABLE workers ADD COLUMN last_login datetime AFTER active;
ALTER TABLE workers ADD COLUMN last_share datetime AFTER last_login;

DROP TABLE ip_addresses;
CREATE TABLE ip_addresses (
	miner_id		int				UNSIGNED NOT NULL,

	ip_address		varchar(40)		NOT NULL,
	active			bool			NOT NULL,
	last_share		datetime		NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_ip_addresses_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	PRIMARY KEY (miner_id, ip_address),
	INDEX idx_ip_addresses_miner_id (miner_id),
	INDEX idx_ip_addresses_last_share (last_share)
);

ALTER TABLE rounds ADD COLUMN worker_id int	UNSIGNED AFTER miner_id;
ALTER TABLE rounds ADD INDEX idx_rounds_worker_id (worker_id);
ALTER TABLE rounds ADD CONSTRAINT fk_rounds_worker_id FOREIGN KEY (worker_id) REFERENCES workers(id);
ALTER TABLE shares ADD COLUMN worker_id int	UNSIGNED AFTER miner_id;
ALTER TABLE shares ADD INDEX idx_shares_worker_id (worker_id);
ALTER TABLE shares ADD CONSTRAINT fk_shares_worker_id FOREIGN KEY (worker_id) REFERENCES workers(id);