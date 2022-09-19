SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE ip_addresses;
TRUNCATE workers;
TRUNCATE miners;
SET FOREIGN_KEY_CHECKS = 1;

ALTER TABLE miners DROP INDEX idx_miners_chain_id_address, ADD UNIQUE INDEX idx_uq_miners_chain_id_address (chain_id, address);
ALTER TABLE workers DROP INDEX idx_workers_miner_id_name, ADD UNIQUE INDEX idx_uq_workers_miner_id_name (miner_id, name);

INSERT INTO 
	miners(chain_id, address, active, recipient_fee_percent) 
VALUES 
	("BTC", "bc1qf4aatnyyxldwhvnaa8fz5gsxq5ceu85lfgrpw6", true, 50),
	("BTC", "16CRhKimYsAy9wXZRXfDdockHcNx3s2h2D", true, 50);