ALTER TABLE raw_blocks ADD COLUMN hash varchar(100)	NOT NULL DEFAULT "" AFTER chain_id;
