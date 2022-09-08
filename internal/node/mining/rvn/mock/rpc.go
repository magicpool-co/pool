package mock

import (
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func GetBlockchainInfo() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"chain":"main","blocks":2432134,"headers":2432134,"bestblockhash":"000000000000916dbcc9c97a768cdb3a065a0b63f2e4fbb073f6b25d8f81e6e9","difficulty":38360.35226237473,"difficulty_algorithm":"DGW-180","mediantime":1661972791,"verificationprogress":0.9999997258213138,"chainwork":"000000000000000000000000000000000000000000000035c63bd0647cdf8f71","size_on_disk":30359747406,"pruned":false,"softforks":[],"bip9_softforks":{"assets":{"status":"active","startTime":1540944000,"timeout":1572480000,"since":435456},"messaging_restricted":{"status":"active","startTime":1578920400,"timeout":1610542800,"since":1092672},"transfer_script":{"status":"active","startTime":1588788000,"timeout":1620324000,"since":1225728},"enforce":{"status":"active","startTime":1593453600,"timeout":1624989600,"since":1306368},"coinbase":{"status":"active","startTime":1597341600,"timeout":1628877600,"since":1409184}},"warnings":""}`))
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

	return rpc.NewResponseFromJSON(nil, []byte(`{"capabilities":["proposal"],"version":805306368,"rules":["assets","messaging_restricted","transfer_script","enforce_value","coinbase"],"vbavailable":{},"vbrequired":0,"previousblockhash":"000000000000916dbcc9c97a768cdb3a065a0b63f2e4fbb073f6b25d8f81e6e9","transactions":[],"coinbaseaux":{"flags":""},"coinbasevalue":250000000000,"longpollid":"000000000000916dbcc9c97a768cdb3a065a0b63f2e4fbb073f6b25d8f81e6e9423800","target":"000000000001b7c9000000000000000000000000000000000000000000000000","mintime":1661972792,"mutable":["time","transactions","prevblock"],"noncerange":"00000000ffffffff","sigoplimit":80000,"sizelimit":8000000,"weightlimit":8000000,"curtime":1661973426,"bits":"1d00ffff","height":2432135,"default_witness_commitment":"6a24aa21a9ede2f61c3f71d1defd3fa999dfa36953755c690689799962b48bebd836974e8cf9"}`))
}

func SubmitBlock(hostID, block string) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`""`))
}

func SendRawTransaction(tx string) *rpc.Response {
	return nil
}
