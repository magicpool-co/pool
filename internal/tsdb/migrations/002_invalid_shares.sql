ALTER TABLE rounds ADD COLUMN invalid_shares double NOT NULL DEFAULT 0 AFTER rejected_shares;
ALTER TABLE global_shares ADD COLUMN invalid_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
ALTER TABLE miner_shares ADD COLUMN invalid_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
ALTER TABLE worker_shares ADD COLUMN invalid_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
