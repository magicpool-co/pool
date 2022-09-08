ALTER TABLE ip_addresses DROP PRIMARY KEY, ADD COLUMN id bigint UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY FIRST;
ALTER TABLE ip_addresses ADD UNIQUE INDEX idx_uq_ip_addresses_miner_id_ip_address(miner_id, ip_address);
