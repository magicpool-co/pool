CREATE TABLE balance_sums (
	miner_id		int				UNSIGNED NOT NULL,
	chain_id		varchar(4)		NOT NULL,

	immature_value	decimal(25,0),
	mature_value	decimal(25,0),

	created_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		datetime		NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_balance_sums_chain_id
	FOREIGN KEY (chain_id)			REFERENCES	chains(id),
	CONSTRAINT fk_balance_sums_miner_id
	FOREIGN KEY (miner_id)			REFERENCES	miners(id),

	PRIMARY KEY (miner_id, chain_id),
	INDEX idx_balance_sums_chain_id (chain_id),
	INDEX idx_balance_sums_miner_id (miner_id),
	INDEX idx_balance_sums_miner_id_chain_id (miner_id, chain_id)
);

INSERT INTO balance_sums
WITH cte AS (
    SELECT
    	miner_id, 
    	chain_id, 
    	sum(value) immature_value, 
    	0 mature_value
    FROM balance_inputs
    WHERE
    	mature = FALSE
    GROUP BY miner_id, chain_id
    UNION ALL
    SELECT
    	miner_id,
    	chain_id,
    	0 immature_value,
    	sum(value) mature_value
    FROM balance_inputs
    WHERE
    	pending = TRUE
    AND
    	mature = TRUE
    GROUP BY miner_id, chain_id
    UNION ALL
    SELECT
    	miner_id,
    	chain_id,
    	0 immature_value,
    	sum(value) mature_value
    FROM balance_outputs
    WHERE
    	spent = FALSE
    GROUP BY miner_id, chain_id
) SELECT
      miner_id,
      chain_id,
      SUM(immature_value) immature_value,
      SUM(mature_value) mature_value,
      CURRENT_TIMESTAMP created_at,
      CURRENT_TIMESTAMP updated_at
FROM cte
GROUP BY miner_id, chain_id
