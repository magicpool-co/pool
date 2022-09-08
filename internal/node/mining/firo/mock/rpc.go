package mock

import (
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func GetBlockchainInfo() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"chain":"main","blocks":529453,"headers":529453,"bestblockhash":"946343845819c5492e393f2884c7eef8d7da1f45e10dafd25e3c4a48ef16753e","difficulty":2059.129661014917,"mediantime":1661919923,"verificationprogress":0.9999953468821784,"chainwork":"000000000000000000000000000000000000000000000000e31768e880e12a60","pruned":false,"softforks":[{"id":"bip34","version":2,"reject":{"status":true}},{"id":"bip66","version":3,"reject":{"status":false}},{"id":"bip65","version":4,"reject":{"status":false}}],"bip9_softforks":{"csv":{"status":"failed","startTime":1462060800,"timeout":1493596800,"since":34272},"segwit":{"status":"failed","startTime":1479168000,"timeout":1510704000,"since":62496}}}`))
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

func GetSpecialTxesCoinbase(hash string) *rpc.Response {
	return nil
}

func GetBlockTemplate() *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`{"capabilities":["proposal"],"version":536875008,"rules":[],"vbavailable":{},"vbrequired":0,"previousblockhash":"946343845819c5492e393f2884c7eef8d7da1f45e10dafd25e3c4a48ef16753e","transactions":[{"data":"0100000001a7cb692ddeaec85ebd9d72a96d26d9190d7539e550b12f5c7c52509c64619056000000006b483045022100f3087671ebcebe58774418a6760ed67f3e504e6a2c5b595fd5accdda281a74180220361618957e4bcda3ca909136cc76e15fcfd7b16e04970b027d1f27284990f57f012102b55e6a43251ef2ab675e50eb0ea5f8df618e0dcbc20f42e2cf9659ef6628a13ffeffffff027b5b4f26000000001976a914520229e7cbfcaf9749af4370f25745cee44786ce88acf6ae6d07000000001976a91468644c7691b4ede5e86003e866ad1b6a1c26330988ac2c140800","txid":"7f66290798ec09f309979f32d877d77728f42f74cc026176f7fb3ac555311571","hash":"7f66290798ec09f309979f32d877d77728f42f74cc026176f7fb3ac555311571","depends":[],"fee":22600,"sigops":8,"weight":904}],"coinbaseaux":{"flags":""},"coinbasevalue":156380530,"longpollid":"946343845819c5492e393f2884c7eef8d7da1f45e10dafd25e3c4a48ef16753e20679","target":"00000000001fd399000000000000000000000000000000000000000000000000","mintime":1661919924,"mutable":["time","transactions","prevblock"],"noncerange":"00000000ffffffff","sigoplimit":400000,"sizelimit":2000000,"weightlimit":2000000,"curtime":1661921270,"bits":"1d00ffff","height":529454,"znode":[{"payee":"a12es1tiXQ9T1L7Lyi6VNKJ7JRrsa7KcgN","script":"76a9140370c5a1974d1409f5e86fba3bfcb3430f51c82388ac","amount":312500000}],"znode_payments_started":true,"znode_payments_enforced":true,"coinbase_payload":"02002e140800b779bd451584eb9b6b13b6d0989c7328e50eba1ca99ef9b4ab0d560ad6f6362e32b51e115296b58b391b778c16700daa198aca4158e9eef32622c7842c6036be","pprpcheader":"a0407a9b188f31150334ec45180882d19eb1bb98e44a3e12bdf1ebd2e319e32d","pprpcepoch":407}`))
}

func SubmitBlock(hostID, block string) *rpc.Response {
	return rpc.NewResponseFromJSON(nil, []byte(`""`))
}

func SendRawTransaction(tx string) *rpc.Response {
	return nil
}
