package flux

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://explorer.runonflux.io/tx/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.runonflux.io/address/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	type response struct {
		Balance uint64 `json:"balanceSat"`
	}

	url := "https://explorer.runonflux.io/api/addr/" + node.address + "/?noTxList=1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	var output *response
	if err := json.NewDecoder(res.Body).Decode(&output); err != nil {
		return nil, err
	} else if res == nil {
		return nil, fmt.Errorf("no response found for %s", node.address)
	}

	return new(big.Int).SetUint64(output.Balance), nil
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	const feeRate = 0
	const expiryHeight = 21

	height, syncing, err := node.GetStatus()
	if err != nil {
		return "", "", err
	} else if syncing {
		return "", "", fmt.Errorf("node is syncing")
	}

	baseTx := btctx.NewTransaction(txVersion, 0, node.prefixP2PKH, node.prefixP2SH, false)
	baseTx.SetVersionMask(versionMask)
	baseTx.SetVersionGroupID(versionGroupID)
	baseTx.SetExpiryHeight(uint32(height + expiryHeight))

	rawTx, err := btctx.GenerateRawTx(baseTx, inputs, outputs, feeRate)
	if err != nil {
		return "", "", err
	}

	rawTxSerialized, err := rawTx.Serialize(nil)
	if err != nil {
		return "", "", err
	}

	tx, err := node.signRawTransaction(hex.EncodeToString(rawTxSerialized), node.wif)
	if err != nil {
		return "", "", err
	}
	txid := btctx.CalculateTxID(tx)

	return txid, tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}
