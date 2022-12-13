package kas

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

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
	balance, err := node.getBalanceByAddress(node.address)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetUint64(balance), nil
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	type apiResponse struct {
		SubnetworkID            string   `json:"subnetwork_id"`
		TransactionID           string   `json:"transaction_id"`
		Hash                    string   `json:"hash"`
		Mass                    string   `json:"mass"`
		BlockHashes             []string `json:"block_hash"`
		BlockTime               int64    `json:"block_time"`
		IsAccepted              bool     `json:"is_accepted"`
		AcceptingBlockHash      string   `json:"accepting_block_hash"`
		AcceptingBlockBlueScore int64    `json:"accepting_block_blue_score"`
	}

	url := fmt.Sprintf("https://api.kaspa.org/transactions/%s?inputs=false&outputs=false", txid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: time.Duration(3 * time.Second),
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data := new(apiResponse)
	err = json.NewDecoder(res.Body).Decode(data)
	if err != nil {
		return nil, err
	} else if data.AcceptingBlockBlueScore < 1 {
		return nil, nil
	}

	txRes := &types.TxResponse{
		Hash:        txid,
		BlockNumber: uint64(data.AcceptingBlockBlueScore),
		Confirmed:   data.IsAccepted,
	}

	return txRes, nil
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
