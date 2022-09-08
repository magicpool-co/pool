ALTER TABLE blocks DROP PRIMARY KEY, ADD COLUMN id bigint UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY FIRST;
ALTER TABLE blocks ADD UNIQUE INDEX idx_uq_blocks_end_time_chain_id_period(end_time, chain_id, period);

ALTER TABLE rounds DROP PRIMARY KEY, ADD COLUMN id bigint UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY FIRST;
ALTER TABLE rounds ADD UNIQUE INDEX idx_uq_rounds_end_time_chain_id_period(end_time, chain_id, period);

ALTER TABLE global_shares DROP PRIMARY KEY, ADD COLUMN id bigint UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY FIRST;
ALTER TABLE global_shares ADD UNIQUE INDEX idx_uq_global_shares_end_time_chain_id_period(end_time, chain_id, period);

ALTER TABLE miner_shares DROP PRIMARY KEY, ADD COLUMN id bigint UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY FIRST;
ALTER TABLE miner_shares ADD UNIQUE INDEX idx_miner_shares_miner_id_end_time_chain_id_period(miner_id, end_time, chain_id, period);

ALTER TABLE worker_shares DROP PRIMARY KEY, ADD COLUMN id bigint UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY FIRST;
ALTER TABLE worker_shares ADD UNIQUE INDEX idx_worker_shares_worker_id_end_time_chain_id_period(worker_id, end_time, chain_id, period);
