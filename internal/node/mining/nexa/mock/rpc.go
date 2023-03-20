package mock

import (
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func GetBlockchainInfo() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"chain":"nexa","blocks":231624,"headers":231624,"bestblockhash":"59744b34393309e31053add7ef6b5a15d75f22eac187f08c3a759bf880be8f0a","difficulty":376762.29308303,"mediantime":1679022934,"verificationprogress":1,"initialblockdownload":false,"chainwork":"00000000000000000000000000000000000000000000000362d815a34ce1b373","coinsupply":231624000000000,"size_on_disk":1266797956,"pruned":false,"softforks":[],"bip9_softforks":{},"bip135_forks":{}}`))
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

func GetMiningCandidate() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"id":1,"headerCommitment":"f3d5cec2ff33f910d3a534dbb4deab4185fc3f60b37872bb7b8863acb8df8a02","nBits":"1a2c8400"}`))
}

func SubmitMiningSolution(hostID string, data map[string]interface{}) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`""`))
}

func SendRawTransaction(tx string) *rpc.Response {
	return nil
}
