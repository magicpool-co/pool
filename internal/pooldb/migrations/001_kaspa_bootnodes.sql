INSERT INTO chains(id, mineable, switchable, payable) VALUES 
	("KAS", true, true, true);

INSERT INTO nodes (url, chain_id, region, mainnet, enabled, backup, active, synced, height) VALUES 
	("node-0.kas.eu-west-1.privatemagicpool.co", "KAS", "eu-west-1", true, true, true, false, false, 0),
	("node-0.kas.us-west-2.privatemagicpool.co", "KAS", "us-west-2", true, false, false, false, false, 0),
	("node-1.kas.us-west-2.privatemagicpool.co", "KAS", "us-west-2", true, false, false, false, false, 0),
	("node-0.kas.eu-central-1.privatemagicpool.co", "KAS", "eu-central-1", true, false, false, false, false, 0),
	("node-1.kas.eu-central-1.privatemagicpool.co", "KAS", "eu-central-1", true, false, false, false, false, 0),
	("node-0.kas.ap-southeast-1.privatemagicpool.co", "KAS", "ap-southeast-1", true, false, false, false, false, 0),
	("node-1.kas.ap-southeast-1.privatemagicpool.co", "KAS", "ap-southeast-1", true, false, false, false, false, 0);
