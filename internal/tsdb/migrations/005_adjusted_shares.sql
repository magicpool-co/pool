ALTER TABLE global_shares ADD COLUMN accepted_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER accepted_shares;
ALTER TABLE miner_shares ADD COLUMN accepted_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER accepted_shares;
ALTER TABLE worker_shares ADD COLUMN accepted_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER accepted_shares;

ALTER TABLE global_shares ADD COLUMN rejected_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
ALTER TABLE miner_shares ADD COLUMN rejected_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
ALTER TABLE worker_shares ADD COLUMN rejected_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;

ALTER TABLE global_shares ADD COLUMN invalid_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
ALTER TABLE miner_shares ADD COLUMN invalid_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
ALTER TABLE worker_shares ADD COLUMN invalid_adjusted_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
