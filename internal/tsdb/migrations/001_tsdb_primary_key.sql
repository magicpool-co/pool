ALTER TABLE blocks DROP COLUMN id, ADD PRIMARY KEY (end_time, chain_id, period);
ALTER TABLE blocks DROP INDEX idx_uq_blocks_end_time_chain_id_period;

ALTER TABLE rounds DROP COLUMN id, ADD PRIMARY KEY (end_time, chain_id, period);
ALTER TABLE rounds DROP INDEX idx_uq_rounds_end_time_chain_id_period;

ALTER TABLE global_shares DROP COLUMN id, ADD PRIMARY KEY (end_time, chain_id, period);
ALTER TABLE global_shares DROP INDEX idx_uq_global_shares_end_time_chain_id_period;

ALTER TABLE miner_shares DROP COLUMN id, ADD PRIMARY KEY (end_time, miner_id, chain_id, period);
ALTER TABLE miner_shares DROP INDEX idx_miner_shares_miner_id_end_time_chain_id_period;

ALTER TABLE worker_shares DROP COLUMN id, ADD PRIMARY KEY (end_time, worker_id, chain_id, period);
ALTER TABLE worker_shares DROP INDEX idx_worker_shares_worker_id_end_time_chain_id_period;
