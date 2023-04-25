ALTER TABLE worker_shares ADD COLUMN reported_hashrate double NOT NULL DEFAULT 0 AFTER avg_hashrate;
ALTER TABLE miner_shares ADD COLUMN reported_hashrate double NOT NULL DEFAULT 0 AFTER avg_hashrate;
ALTER TABLE global_shares ADD COLUMN reported_hashrate double NOT NULL DEFAULT 0 AFTER avg_hashrate;
