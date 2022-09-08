package mock

func GetAddressFromErgoTree() []byte {
	return nil
}

func GetInfo() []byte {
	return []byte(`{"currentTime":1662401307197,"network":"mainnet","name":"ergo-mainnet-4.0.35","stateType":"utxo","difficulty":2465921113260032,"bestFullHeaderId":"436c3274b70a2aa7d559a7057016a6ee589a5a473c599d6d59183e2e7b688cde","bestHeaderId":"436c3274b70a2aa7d559a7057016a6ee589a5a473c599d6d59183e2e7b688cde","peersCount":30,"unconfirmedCount":64,"appVersion":"4.0.35","stateRoot":"89f64e6b20f6174f3ff8d5e45a5b08ee6abd104dbcc403c8eed9209091946b7219","genesisBlockId":"b0244dfc267baca974a4caee06120321562784303a8a688976ae56170e4d175b","previousFullHeaderId":"91a2af3cec91044ac336b610162f57198fad1e59d3fe9a75aab1170995532022","fullHeight":832295,"headersHeight":832295,"stateVersion":"436c3274b70a2aa7d559a7057016a6ee589a5a473c599d6d59183e2e7b688cde","fullBlocksScore":1028799373229971996672,"maxPeerHeight":832295,"launchTime":1662398449903,"lastSeenMessageTime":1662401284808,"eip27Supported":true,"headersScore":1028799373229971996672,"parameters":{"outputCost":173,"tokenAccessCost":100,"maxBlockCost":8001091,"height":831488,"maxBlockSize":1271009,"dataInputCost":100,"blockVersion":2,"inputCost":2000,"storageFeeFactor":1250000,"minValuePerByte":360},"isMining":true}`)
}

func GetWalletBalances() []byte {
	return nil
}

func GetBlockAtHeight() []byte {
	return nil
}

func GetBlock() []byte {
	return nil
}

func GetRewardAddress() []byte {
	return []byte(`{"rewardAddress":"88dhgzEuTXaUw7sx6Ku3iF8TGjZr4bjZvJV8rR3gJYAsYa28fYGMpbu4WW8FDxJ51wqZungH47zrLCrs"}`)
}

func GetMiningCandidate() []byte {
	return []byte(`{"msg":"4bf554fc90d919113ec4c8b26ca093374181e2016bf0336503e80c4335125467","b":46956931677443280224521137779538216457838764113420434010722637,"h":832296,"pk":"03934a363523b2d8678362801c8191d67163834859923acbfb6ab63194e526d18e"}`)
}

func GetWalletStatus() []byte {
	return []byte(`{"isInitialized":true,"isUnlocked":false,"changeAddress":"","walletHeight":832295,"error":""}`)
}
