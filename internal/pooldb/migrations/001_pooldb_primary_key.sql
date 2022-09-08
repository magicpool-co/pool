ALTER TABLE ip_addresses DROP COLUMN id, ADD PRIMARY KEY (miner_id, ip_address);
ALTER TABLE ip_addresses DROP INDEX idx_uq_ip_addresses_miner_id_ip_address;
