package firo

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
	return "https://explorer.firo.org/tx/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.firo.org/address/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	type response struct {
		Balance uint64 `json:"balanceSat"`
	}

	url := "https://explorer.firo.org/insight-api-zcoin/addr/" + node.address + "/?noTxList=1"
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
	tx, err := node.getRawTransaction(txid)
	if err != nil {
		return nil, err
	}

	var height uint64
	var confirmed bool
	if tx.Height > 0 && tx.Confirmations > 0 {
		confirmed = true
		height = uint64(tx.Height)
	}

	res := &types.TxResponse{
		Hash:        txid,
		BlockNumber: height,
		Confirmed:   confirmed,
	}

	return res, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	// @TODO: figure out proper fee rate
	const feeRate = 2000

	baseTx := btctx.NewTransaction(txVersion, 0, node.prefixP2PKH, node.prefixP2SH, false)
	rawTx, err := btctx.GenerateTx(node.privKey, baseTx, inputs, outputs, feeRate)
	if err != nil {
		return "", "", err
	}
	tx := hex.EncodeToString(rawTx)
	txid := btctx.CalculateTxID(tx)

	return txid, tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("sendrawtransaction", tx)
	if err != nil {
		return "", err
	}

	var txid string
	if err := json.Unmarshal(res.Result, &txid); err != nil {
		return "", err
	}

	return txid, nil
}
