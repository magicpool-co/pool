package swap

import (
	"bytes"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/ethtx"
	"github.com/magicpool-co/pool/types"
)

type FusionClient struct {
	url string
}

func NewFusionClient() *FusionClient {
	client := &FusionClient{
		url: "https://fusion.runonflux.io",
	}

	return client
}

// http helpers

func (c *FusionClient) do(method, path string, body, target interface{}) error {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, c.url+path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	client := http.Client{Timeout: time.Duration(3 * time.Second)}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	var output *FusionResponse
	if err := json.NewDecoder(res.Body).Decode(&output); err != nil {
		return err
	}

	switch output.Status {
	case "success":
		return json.Unmarshal(output.Data, target)
	case "error":
		var errBody *FusionError
		if err := json.Unmarshal(output.Data, &errBody); err != nil {
			return fmt.Errorf("failed request error handling: %v", err)
		}
		return fmt.Errorf("failed request: %s: %s", errBody.Name, errBody.Message)
	default:
		return fmt.Errorf("unknown response status: %s", output.Status)
	}
}

func (c *FusionClient) getSwapInfo() (*FusionSwapInfo, error) {
	var info *FusionSwapInfo
	err := c.do("GET", "/swap/info", nil, &info)

	return info, err
}

func (c *FusionClient) getSwapDetail(txid string) (*FusionSwapDetail, error) {
	var detail *FusionSwapDetail
	err := c.do("GET", "/swap/detail?txid="+txid, nil, &detail)

	return detail, err
}

func (c *FusionClient) reserveSwap(fromChain, fromAddr, toChain, toAddr string) (*FusionSwapReservation, error) {
	data := map[string]interface{}{
		"chainFrom":   fromChain,
		"chainTo":     toChain,
		"addressFrom": fromAddr,
		"addressTo":   toAddr,
	}

	var reservation *FusionSwapReservation
	err := c.do("POST", "/swap/reserve", data, &reservation)

	return reservation, err
}

func (c *FusionClient) createSwap(fromChain, fromAddr, toChain, toAddr, txidFrom string, amount float64) (*FusionSwapDetail, error) {
	data := map[string]interface{}{
		"amountFrom":  amount,
		"chainFrom":   fromChain,
		"chainTo":     toChain,
		"addressFrom": fromAddr,
		"addressTo":   toAddr,
		"txidFrom":    txidFrom,
	}

	var detail *FusionSwapDetail
	err := c.do("POST", "/swap/create", data, &detail)

	return detail, err
}

// swaps

func (c *FusionClient) InitiateSwapFromFlux(fluxNode types.PayoutNode, bscAddress string, inputs []*types.TxInput) (string, error) {
	amount := new(big.Int)
	for _, input := range inputs {
		amount.Add(amount, input.Value)
	}
	amountFloat64 := float64(amount.Uint64()) / 1e8

	info, err := c.getSwapInfo()
	if err != nil {
		return "", err
	}

	var depositAddress string
	for _, chain := range info.SwapAddresses {
		switch chain.Chain {
		case "main":
			depositAddress = chain.Address
		}
	}
	if depositAddress == "" {
		return "", fmt.Errorf("deposit address not found")
	}

	fees := float64(info.Fees.Swap.BSC + 1)
	if amountFloat64-fees <= 0 {
		return "", fmt.Errorf("not enough value to cover fees")
	}

	_, err = c.reserveSwap("main", fluxNode.Address(), "bsc", bscAddress)
	if err != nil {
		return "", err
	}

	outputs := []*types.TxOutput{
		&types.TxOutput{
			Address: depositAddress,
			Value:   amount,
		},
	}

	tx, err := fluxNode.CreateTx(inputs, outputs)
	if err != nil {
		return "", err
	}

	txid, err := fluxNode.BroadcastTx(tx)
	if err != nil {
		return "", err
	}

	detail, err := c.createSwap("main", fluxNode.Address(), "bsc", bscAddress, txid, amountFloat64)
	if err != nil {
		return "", err
	}

	return detail.TxIDFrom, nil
}

func (c *FusionClient) InitiateSwapFromBSC(bscNode types.PayoutNode, fluxAddress string, amount *big.Int) (string, error) {
	const contractAddress = "0xaff9084f2374585879e8b434c399e29e80cce635"

	info, err := c.getSwapInfo()
	if err != nil {
		return "", err
	}

	var depositAddress string
	for _, chain := range info.SwapAddresses {
		switch chain.Chain {
		case "bsc":
			depositAddress = chain.Address
		}
	}
	if depositAddress == "" {
		return "", fmt.Errorf("deposit address not found")
	}

	depositAddressBytes, err := common.HexToBytes(depositAddress)
	if err != nil {
		return "", err
	}

	amountFloat64 := float64(amount.Uint64()) / 1e18
	fees := float64(info.Fees.Swap.FLUX)
	if amountFloat64-fees <= 0 {
		return "", fmt.Errorf("not enough value to cover fees")
	}

	_, err = c.reserveSwap("bsc", bscNode.Address(), "main", fluxAddress)
	if err != nil {
		return "", err
	}

	data := ethtx.GenerateContractData("transfer(address,uint256)", depositAddressBytes, amount.Bytes())
	inputs := []*types.TxInput{&types.TxInput{Value: new(big.Int), Data: data}}
	outputs := []*types.TxOutput{&types.TxOutput{Address: contractAddress, Value: new(big.Int)}}
	tx, err := bscNode.CreateTx(inputs, outputs)
	if err != nil {
		return "", err
	}

	txid, err := bscNode.BroadcastTx(tx)
	if err != nil {
		return "", err
	}

	detail, err := c.createSwap("bsc", bscNode.Address(), "flux", fluxAddress, txid, amountFloat64)
	if err != nil {
		return "", err
	}

	return detail.TxIDFrom, nil
}

func (c *FusionClient) FinalizeSwap(txid string) error {
	detail, err := c.getSwapDetail(txid)
	if err != nil {
		return nil
	}

	switch strings.ToLower(detail.Status) {
	case "finished":
		return nil
	case "new", "waiting", "confirming", "exchanging", "hold":
		return ErrSwapNotReady
	case "expired":
		return ErrSwapExpired
	case "dust":
		return ErrSwapDust
	default:
		return fmt.Errorf("err unkonw status %s", detail.Status)
	}
}
