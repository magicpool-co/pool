package kas

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"google.golang.org/protobuf/proto"

	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
	"github.com/magicpool-co/pool/pkg/crypto/tx/kastx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://explorer.kaspa.org/txs/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.kaspa.org/addresses/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	return new(big.Int), nil
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	const feePerInput uint64 = 10000

	txBytes, err := kastx.GenerateTx(node.privKey, inputs, outputs, node.prefix, feePerInput)
	if err != nil {
		return "", "", err
	}

	txHex := hex.EncodeToString(txBytes)
	txid := kastx.CalculateTxID(txHex)

	return txid, txHex, nil
}

func (node Node) BroadcastTx(txHex string) (string, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return "", err
	}

	tx := new(protowire.RpcTransaction)
	err = proto.Unmarshal(txBytes, tx)
	if err != nil {
		return "", err
	}

	t, _ := json.Marshal(tx)
	fmt.Println(string(t))

	return node.submitTransaction(tx)
}
