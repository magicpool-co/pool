ALTER TABLE miners ADD COLUMN recipient_fee_percent int UNSIGNED AFTER address;

CREATE TABLE balance_inputs (
	id				bigint			UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	round_id		int         	UNSIGNED NOT NULL,
	chain_id		varchar(4)		NOT NULL,
	miner_id		int				UNSIGNED NOT NULL,
	
	output_balance_id	bigint UNSIGNED,

	value			decimal(25,0)	NOT NULL,
	pending			bool			NOT NULL,

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_balance_inputs_round_id
	FOREIGN KEY (round_id)			REFERENCES	rounds(id),
	CONSTRAINT fk_balance_inputs_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_balance_inputs_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	INDEX idx_balance_inputs_round_id (round_id),
	INDEX idx_balance_inputs_chain_id (chain_id),
	INDEX idx_balance_inputs_miner_id (miner_id),
	INDEX idx_balance_inputs_output_balance_id (output_balance_id)
);