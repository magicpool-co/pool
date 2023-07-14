ALTER TABLE worker_shares DROP COLUMN invalid_adjusted_shares;
ALTER TABLE miner_shares DROP COLUMN invalid_adjusted_shares;
ALTER TABLE global_shares DROP COLUMN invalid_adjusted_shares;

ALTER TABLE worker_shares DROP COLUMN rejected_adjusted_shares;
ALTER TABLE miner_shares DROP COLUMN rejected_adjusted_shares;
ALTER TABLE global_shares DROP COLUMN rejected_adjusted_shares;

ALTER TABLE worker_shares DROP COLUMN accepted_adjusted_shares;
ALTER TABLE miner_shares DROP COLUMN accepted_adjusted_shares;
ALTER TABLE global_shares DROP COLUMN accepted_adjusted_shares;
