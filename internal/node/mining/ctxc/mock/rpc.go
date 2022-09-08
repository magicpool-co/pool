package mock

import (
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func GetTransactionReceiptMany(txids []string) []*rpc.Response {
	return nil
}

func GetBlockByNumberMany(heights []uint64) []*rpc.Response {
	return nil
}

func GetBalance(address string) *rpc.Response {
	return nil
}

func GetChainID() *rpc.Response {
	return nil
}

func GetGasPrice() *rpc.Response {
	return nil
}

func GetPendingNonce(address string) *rpc.Response {
	return nil
}

func GetBlockNumber() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`0x685ddc`))
}

func GetSyncing() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, common.JsonFalse)
}

func GetBlockByNumber(height uint64) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"difficulty":"0x14bc","extraData":"0x436f7274657850504c4e532f326d696e6572735f4555","gasLimit":"0x7a1200","gasUsed":"0x0","hash":"0x43992e4d77ab8dbdeab1a97eb492bb0515b46a1ef6e91e9d73dc3ed914fdcca2","logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","miner":"0xa5f2c7c78e90e01b21d25dc0971b63e7a50f88a6","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","nonce":"0x9dc79f00a56172f7","number":"0x685ddc","parentHash":"0x1e4eca4ad02820d5ec45bb1acae2eee4dcf2a1d7c614e9fcd79223076c404f44","quota":"0x685ddc0000","quotaUsed":"0x7804460","receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","size":"0x301","solution":[13608074,42736631,54825486,99642100,146179831,157659643,165181349,217761886,235165505,273697782,284704189,295862087,324589697,330250357,396051660,426356538,463394489,486745775,488859732,579888629,604722188,611650182,613211719,629234456,654619510,660045480,705977491,725551429,734503893,778986150,825311203,834278032,855007896,930779059,945302433,951450587,981020827,988546049,1007646198,1020525811,1041448089,1066456251],"stateRoot":"0xeb44bfc423d6de33ac8451190191f510ade0243bbadede625baf6e19fffff4d6","supply":"0xa584340003379a24986000","timestamp":"0x630e74d5","totalDifficulty":"0x5c72eb737","transactions":[],"transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","uncles":[]}`))
}

func GetUncleByNumberAndIndex(height, index uint64) *rpc.Response {
	return nil
}

func GetWork() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`["0x3f62b7a620e9695597115ab1d5a219162136e7af6f187796b1a244793d27175d","0x0000000000000000000000000000000000000000000000000000000000000000","0x1000000000000000000000000000000000000000000000000000000000000000","0x685ddc"]`))
}

func SendEstimateGas(from, to string) *rpc.Response {
	return nil
}

func SendSubmitWork(nonce, hash, solution string) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, common.JsonTrue)
}

func SendRawTransaction(tx string) *rpc.Response {
	return nil
}
