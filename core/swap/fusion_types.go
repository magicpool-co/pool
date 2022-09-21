package swap

import (
	"github.com/goccy/go-json"
)

type FusionResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

type FusionError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type FusionSwapFees struct {
	Percentage        float32 `json:"percentage"`
	PremiumPercentage float32 `json:"premiumPercentage"`
	FLUX              int     `json"main"`
	KDA               int     `json"kda"`
	BSC               int     `json"bsc"`
	ETH               int     `json"eth"`
	TRX               int     `json"trx"`
	SOL               int     `json"sol"`
	AVAX              int     `json"avax"`
}

type FusionSwapInfo struct {
	Chains []string `json:"chains"`
	Coins  []struct {
		Coin  string `json:"coin"`
		Chain string `json:"chain"`
	} `json:"coins"`
	Confirmations []struct {
		Confirmations int `json:"confirmations"`
		Amount        int `json:"amount"`
	} `json:"confirmations"`
	Fees struct {
		Snapshot FusionSwapFees `json:"snapshot"`
		Mining   FusionSwapFees `json:"mining"`
		Swap     FusionSwapFees `json:"swap"`
	} `json:"fees"`
	SwapAddresses []struct {
		Chain   string `json:"chain"`
		Address string `json:"address"`
	} `json:"swapAddresses"`
}

type FusionSwapDetail struct {
	ExpectedAmountFrom float32 `json:"expectedAmountFrom"`
	ExpectedAmountTo   float32 `json:"expectedAmountTo"`
	ChainFrom          string  `json:"chainFrom"`
	ChainTo            string  `json:"chainTo"`
	AddressFrom        string  `json:"addressFrom"`
	AddressTo          string  `json:"addressTo"`
	TxIDFrom           string  `json:"txidFrom"`
	ZelID              string  `json:"zelid"`
	Fee                int     `json:"fee"`
	Status             string  `json:"status"`
	Timestamp          int64   `json:"timestamp"`
	ConfsRequired      int     `json:"confsRequired"`
	AmountFrom         float32 `json:"amountFrom"`
	AmountTo           float32 `json:"amountTo"`
	TxIDTo             string  `json:"txidTo"`
}

type FusionSwapReservation struct {
	ChainFrom   string `json:"chainFrom"`
	ChainTo     string `json:"chainTo"`
	AddressFrom string `json:"addressFrom"`
	AddressTo   string `json:"addressTo"`
	ZelID       string `json:"zelid"`
	CreatedAt   string `json:"createdAt"`
}
