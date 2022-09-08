package mock

import (
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func GetBlockchainInfo() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"chain":"main","blocks":1198291,"headers":1198291,"bestblockhash":"00000008705d035f6d48d340dbef3ff05b0de40354c896b91d38fa588f0f096b","difficulty":24426.6318453515,"verificationprogress":0.9999986057034661,"chainwork":"0000000000000000000000000000000000000000000000000000542f27969f95","pruned":false,"size_on_disk":11025831509,"commitments":815732,"valuePools":[{"id":"sprout","monitored":true,"chainValue":19300.55579041,"chainValueZat":1930055579041},{"id":"sapling","monitored":true,"chainValue":77188.75921324,"chainValueZat":7718875921324}],"softforks":[{"id":"bip34","version":2,"enforce":{"status":true,"found":4000,"required":750,"window":4000},"reject":{"status":true,"found":4000,"required":950,"window":4000}},{"id":"bip66","version":3,"enforce":{"status":true,"found":4000,"required":750,"window":4000},"reject":{"status":true,"found":4000,"required":950,"window":4000}},{"id":"bip65","version":4,"enforce":{"status":true,"found":4000,"required":750,"window":4000},"reject":{"status":true,"found":4000,"required":950,"window":4000}}],"upgrades":{"76b809bb":{"name":"Acadia","activationheight":250000,"status":"active","info":"The Zelcash Acadia Update"},"76b809bb":{"name":"Kamiooka","activationheight":372500,"status":"active","info":"Zel Kamiooka Upgrade, PoW change to ZelHash and update for ZelNodes"},"76b809bb":{"name":"Kamata","activationheight":558000,"status":"active","info":"Zel Kamata Upgrade, Deterministic ZelNodes and ZelFlux"},"76b809bb":{"name":"Flux","activationheight":835554,"status":"active","info":"Flux Upgrade, Multiple chains"},"76b809bb":{"name":"Halving","activationheight":1076532,"status":"active","info":"Flux Halving"}},"consensus":{"chaintip":"76b809bb","nextblock":"76b809bb"}}`))
}

func GetBlockHash(height uint64) *rpc.Response {
	return nil
}

func GetBlockHashMany(heights []uint64) []*rpc.Response {
	return nil
}

func GetBlock(hash string) *rpc.Response {
	return nil
}

func GetBlockMany(hashes []string) []*rpc.Response {
	return nil
}

func GetBlockTemplate() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"capabilities":["proposal"],"version":4,"previousblockhash":"00000008705d035f6d48d340dbef3ff05b0de40354c896b91d38fa588f0f096b","finalsaplingroothash":"22758d2483e36192e2ca6beef539801f3d6218eb6e24125da16d9cdc7a8e98d9","transactions":[{"data":"0500000004b6efcdf33c060677e5bbec1b5449d22a9df939a039e22eba3a4bff9e624b04ef000000009be70e63019ce20e63010c38392e3234352e3231322e30411b35261ec3eff1f172954c971af272016be8f7505e2368ece14de7d5745657ef26603acb7fd841cc23770872c7d77096ca90093a500eaf1ed859594d0bccaad638411c9d95c4d055080e90f759e2c0d77bff957725de29538c686887cab17d270791a65b90f3a43315fd58e2e52efcb54ff770babde238c90e89ab05bc7ff25303b721","hash":"56160a307206f123736601bab6929266ed8fd6c7e44e4961e2badad4a8d3d623","depends":[],"fee":0,"sigops":0}],"coinbasetxn":{"data":"0400008085202f89010000000000000000000000000000000000000000000000000000000000000000ffffffff0503d4481200ffffffff04807584df000000001976a914cb27f1b9a165bdcfca1ffb2ba5fa6e2c43eba02c88aca0118721000000001976a9141aacb774eb225461ce8bf37e7843f8e3858923cf88ac601de137000000001976a914d240cdcf21eed393ddec1ffe3b50f36aa05ab10b88ac80461c86000000001976a914064fe4624c2e8e50f78a24154546bfce9860125f88ac00000000000000000000000000000000000000","hash":"49cff17b2aed596003c4dc0d1196525740dc8d6ade9996ff35bc24b88e4562de","depends":[],"fee":0,"sigops":4,"required":true},"longpollid":"00000008705d035f6d48d340dbef3ff05b0de40354c896b91d38fa588f0f096b7305006","target":"0000001576b80000000000000000000000000000000000000000000000000000","mintime":1661920545,"mutable":["time","transactions","prevblock"],"noncerange":"00000000ffffffff","sigoplimit":20000,"sizelimit":2000000,"curtime":1661921213,"bits":"1d00ffff","height":1198292,"miner_reward":3750000000,"basic_zelnode_address":"t1LJeUtGto9svb1WuX3H3z37kNuKgjxxvwz","basic_zelnode_payout":562500000,"cumulus_fluxnode_address":"t1LJeUtGto9svb1WuX3H3z37kNuKgjxxvwz","cumulus_fluxnode_payout":562500000,"super_zelnode_address":"t1d3KYH1MCc2nS2rZ1JiUxZzn6Ct2VyNt3Z","super_zelnode_payout":937500000,"nimbus_fluxnode_address":"t1d3KYH1MCc2nS2rZ1JiUxZzn6Ct2VyNt3Z","nimbus_fluxnode_payout":937500000,"bamf_zelnode_address":"t1JSymVJHhu9aCTA4xqLevAkNNB8Z3tsWiR","bamf_zelnode_payout":2250000000,"stratus_fluxnode_address":"t1JSymVJHhu9aCTA4xqLevAkNNB8Z3tsWiR","stratus_fluxnode_payout":2250000000}`))
}

func SubmitBlock(hostID, block string) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`""`))
}

func SendRawTransaction(tx string) *rpc.Response {
	return nil
}
