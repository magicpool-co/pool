package mock

import (
	"time"

	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func GetBalance(address string) *rpc.Response {
	return nil
}

func GetLatestBlock() *rpc.Response {
	return nil
}

func GetBlockRewardInfo(epochHeight uint64) *rpc.Response {
	return nil
}

func GetBlockRewardInfoMany(epochHeights []uint64) []*rpc.Response {
	return nil
}

func GetBlockByHash(blockHash string) *rpc.Response {
	return nil
}

func GetBlockByHashMany(blockHashes []string) []*rpc.Response {
	return nil
}

func GetBlocksByEpochMany(epochHeights []uint64) []*rpc.Response {
	return nil
}

func GetGasPrice() *rpc.Response {
	return nil
}

func GetPendingNonce(address string) *rpc.Response {
	return nil
}

func GetEpochNumber() *rpc.Response {
	return nil
}

func SendEstimateGas(from, to string) *rpc.Response {
	return nil
}

func SendRawTransaction(tx string) *rpc.Response {
	return nil
}

func MiningSubscribe() chan *rpc.Request {
	ch := make(chan *rpc.Request)
	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 100):
				ch <- rpc.MustNewRequest("mining.notify",
					"0xde5b0ae317379fa03f768eb102fb1e3671c9beaafecc52ae3c24eb77e80d6e03",
					"53013744",
					"0xde5b0ae317379fa03f768eb102fb1e3671c9beaafecc52ae3c24eb77e80d6e03",
					"0x000000044b82fa09b5a52cb98b405447c4a98187eebb22f008d5d64f9c394ae8",
				)
				return
			}
		}
	}()

	return ch
}

func SubmitBlock(hostID, nonce, hash string) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`[true]`))
}
