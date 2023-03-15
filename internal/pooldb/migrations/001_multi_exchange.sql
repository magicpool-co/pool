ALTER TABLE exchange_trades ADD COLUMN step_id tinyint(1) UNSIGNED NOT NULL DEFAULT 0 AFTER stage_id;
ALTER TABLE exchange_trades ADD COLUMN is_market_order bool NOT NULL DEFAULT TRUE AFTER step_id;
ALTER TABLE exchange_trades ADD COLUMN trade_strategy tinyint(1) UNSIGNED NOT NULL DEFAULT 0 AFTER is_market_order;

UPDATE exchange_batches SET status = status + 2 WHERE status > 8;
