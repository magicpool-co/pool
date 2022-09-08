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
	return rpc.NewResponseFromJSON(nil, []byte(`0xeba1d6`))
}

func GetSyncing() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, common.JsonFalse)
}

func GetBlockByNumber(height uint64) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"baseFeePerGas":"0x2aa34af9a","difficulty":"0x2c003048f7fa6d","extraData":"0x617369612d65617374322d3130","gasLimit":"0x1ca35d3","gasUsed":"0x191202c","hash":"0x2a6ccb1eca1e9f73c4ce40bed1d941ab519bae206c9a1044dd05ce08c9c2e044","logsBloom":"","miner":"0xea674fdde714fd979de3edf0f56aa9716b898ec8","mixHash":"0x79d91c70c99525adc9dc1d15cd953a189474234f8e5a068270c5f2ee67248dfe","nonce":"0x55d4e788d7bf7410","number":"0xeba1d6","parentHash":"0x80f8634ffa80699fa05aee16933c86273bada9ffcf8bd6d2a4427f79c2e5be08","receiptsRoot":"0xb9f7c9acfab74b71faf313949e7c931144cfddd3b6e14b70b70c56306c023bb0","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","size":"0x306eb","stateRoot":"0x55109a5872005b5005054c28aaefd3db32ed19843ec3fb74ff2a0e15c58497e5","timestamp":"0x630e6f9b","totalDifficulty":"0xc31854fcadaed5cf964","transactions":[],"transactionsRoot":"0xa7b2f5ecfc754f5648b2c20524f8958133acefad116d619ed58a95a914580bfd","uncles":[]}`))
}

func GetUncleByNumberAndIndex(height, index uint64) *rpc.Response {
	return nil
}

func GetWork() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`["0x48b9e2560c8263614076d943d2c848044604b6e43c2423ed670ea5cc18b6edd8","0xb883d93ea88f16bb9d1318c55fb480d7df858b6dbb77c26804cb7a51030050b8","0x000000007e00000007e00000007e00000007e00000007e00000007e00000007e","0xeba1d6"]`))
}

func SendEstimateGas(from, to string) *rpc.Response {
	return nil
}

func SendSubmitWork(nonce, hash, mixDigest string) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, common.JsonTrue)
}

func SendCall(params []interface{}) *rpc.Response {
	return nil
}

func SendRawTransaction(tx string) *rpc.Response {
	return nil
}
