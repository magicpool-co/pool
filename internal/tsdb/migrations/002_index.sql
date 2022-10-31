ALTER TABLE blocks ADD INDEX idx_blocks_chain_id_period (chain_id, period);

ALTER TABLE global_shares ADD INDEX idx_global_shares_chain_id_period (chain_id, period);
ALTER TABLE global_shares ADD INDEX idx_global_shares_end_time_chain_id_period (end_time, chain_id, period);

ALTER TABLE miner_shares ADD INDEX idx_miner_shares_miner_id_period (miner_id, period);
ALTER TABLE miner_shares ADD INDEX idx_miner_shares_end_time_miner_id_chain_id_period (end_time, miner_id, chain_id, period);

ALTER TABLE worker_shares ADD INDEX idx_worker_shares_worker_id_period (worker_id, period);
ALTER TABLE worker_shares ADD INDEX idx_worker_shares_end_time_worker_id_chain_id_period (end_time, worker_id, chain_id, period);
