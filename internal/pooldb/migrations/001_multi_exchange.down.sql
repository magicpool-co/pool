ALTER TABLE exchange_trades DROP COLUMN exchange_id;
ALTER TABLE exchange_trades DROP COLUMN trade_strategy;
ALTER TABLE exchange_trades DROP COLUMN is_market_order;
ALTER TABLE exchange_trades DROP COLUMN step_id;

UPDATE exchange_batches SET status = status - 2 WHERE status > 8;
