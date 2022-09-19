ALTER TABLE miners DROP INDEX idx_uq_miners_chain_id_address, ADD INDEX idx_miners_chain_id_address (chain_id, address);
ALTER TABLE workers DROP INDEX idx_uq_workers_miner_id_name, ADD INDEX idx_workers_miner_id_name (miner_id, name);
