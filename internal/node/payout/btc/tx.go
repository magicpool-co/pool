package btc

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/blockchair"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/types"
)

const (
	txVersion = 0x1
)

type feeResponse interface {
	getRate() uint64
}

type blockchainInfoFeeResponse struct {
	Limits struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"limits"`
	Regular  int    `json:"regular"`
	Priority uint64 `json:"priority"`
}

func (b *blockchainInfoFeeResponse) getRate() uint64 {
	return b.Priority
}

type bitgoFeeResponse struct {
	FeePerKB   uint64 `json:"feePerKB"`
	NumBlocks  int    `json:"numBlocks"`
	Confidence int    `json:"confidence"`
	Multiplier int    `json:"multiplier"`
}

func (b *bitgoFeeResponse) getRate() uint64 {
	if b.Confidence > 60 {
		return b.FeePerKB / 1000
	}

	return 0
}

type bitcoinerLiveFeeResponse struct {
	Timestamp int `json:"timestamp"`
	Estimates map[int]struct {
		SatsPerVByte uint64 `json:"sats_per_vbyte"`
	} `json:"estimates"`
}

func (b *bitcoinerLiveFeeResponse) getRate() uint64 {
	if val, ok := b.Estimates[30]; ok {
		return val.SatsPerVByte / 1000
	}

	return 0
}

func getFeeRate() (uint64, error) {
	feeSources := map[string]feeResponse{
		"https://api.blockchain.info/mempool/fees":         new(blockchainInfoFeeResponse),
		"https://bitcoiner.live/api/fees/estimates/latest": new(bitgoFeeResponse),
		"https://www.bitgo.com/api/v2/btc/tx/fee":          new(bitcoinerLiveFeeResponse),
	}

	for url, data := range feeSources {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		client := http.Client{
			Timeout: time.Duration(3 * time.Second),
		}
		res, err := client.Do(req)
		if err != nil {
			continue
		}

		defer res.Body.Close()
		err = json.NewDecoder(res.Body).Decode(data)
		if err != nil || data.getRate() == 0 {
			continue
		}

		return data.getRate(), nil
	}

	return 0, fmt.Errorf("unable to find BTC fee rate")
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	tx, err := blockchair.New(node.blockchairKey).GetTxBTC(txid)
	if err != nil {
		return nil, err
	}

	outputs := make([]*types.UTXOResponse, len(tx.Outputs))
	for i, output := range tx.Outputs {
		outputs[i] = &types.UTXOResponse{
			Hash:    output.TransactionHash,
			Index:   output.Index,
			Value:   output.Value,
			Address: output.Recipient,
		}
	}

	confirmed := false
	var height uint64 = 0
	if tx.Tx.BlockID != -1 {
		confirmed = true
		height = uint64(tx.Tx.BlockID)
	}

	res := &types.TxResponse{
		Hash:        tx.Tx.Hash,
		BlockNumber: height,
		Value:       new(big.Int).SetUint64(tx.Tx.OutputTotal),
		Fee:         new(big.Int).SetUint64(tx.Tx.Fee),
		FeeBalance:  new(big.Int),
		Confirmed:   confirmed,
		Outputs:     outputs,
	}

	return res, nil
}

func (node Node) GetBalance(addr string) (*big.Int, error) {
	address, err := blockchair.New(node.blockchairKey).GetAddressBTC(addr)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetUint64(address.Balance), nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, error) {
	feeRate, err := getFeeRate()
	if err != nil {
		return "", err
	}

	baseTx := btctx.NewTransaction(txVersion, 0, node.prefixP2PKH, node.prefixP2SH, true)
	rawTx, err := btctx.GenerateTx(node.privKey, baseTx, inputs, outputs, feeRate)
	if err != nil {
		return "", err
	}
	tx := hex.EncodeToString(rawTx)

	return tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return blockchair.New(node.blockchairKey).BroadcastTxBTC(tx)
}
