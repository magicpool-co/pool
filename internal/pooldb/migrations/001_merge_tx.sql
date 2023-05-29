ALTER TABLE balance_outputs ADD COLUMN out_merge_transaction_id BIGINT UNSIGNED AFTER out_payout_id;
ALTER TABLE balance_outputs ADD INDEX idx_balance_outputs_out_merge_transaction_id (out_merge_transaction_id);
ALTER TABLE balance_outputs ADD CONSTRAINT fk_balance_outputs_out_merge_transaction_id 
	FOREIGN KEY (out_merge_transaction_id) REFERENCES transactions(id);
