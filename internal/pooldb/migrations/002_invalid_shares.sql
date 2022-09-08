ALTER TABLE rounds ADD COLUMN invalid_shares bigint UNSIGNED NOT NULL DEFAULT 0 AFTER rejected_shares;
