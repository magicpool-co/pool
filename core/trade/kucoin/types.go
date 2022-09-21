package kucoin

import (
	"github.com/goccy/go-json"
)

type Response struct {
	Code    string          `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type PaginatedResponse struct {
	CurrentPage int64           `json:"currentPage"`
	PageSize    int64           `json:"pageSize"`
	TotalNum    int64           `json:"totalNum"`
	TotalPage   int64           `json:"totalPage"`
	Items       json.RawMessage `json:"items"`
}

type Status struct {
	Message string `json:"msg"`
	Status  string `json:"status"`
}

type Symbol struct {
	Sequence    string `json:"sequence"`
	Price       string `json:"price"`
	Size        string `json:"size"`
	BestBid     string `json:"bestBid"`
	BestBidSize string `json:"bestBidSize"`
	BestAsk     string `json:"bestAsk"`
	BestAskSize string `json:"bestAskSize"`
	Time        int64  `json:"time"`
}

type Chain struct {
	ChainName         string `json:"chainName"`
	WithdrawalMinSize string `json:"withdrawalMinSize"`
	WithdrawalMinFee  string `json:"withdrawalMinFee"`
	IsWithdrawEnabled bool   `json:"isWithdrawEnabled"`
	IsDepositEnabled  bool   `json:"isDepositEnabled"`
	Confirms          int64  `json:"confirms"`
	ContractAddress   string `json:"contractAddress"`
}

type Currency struct {
	Name            string   `json:"name"`
	Currency        string   `json:"currency"`
	FullName        string   `json:"fullName"`
	Precision       uint8    `json:"precision"`
	Confirms        int64    `json:"confirms"`
	ContractAddress string   `json:"contractAddress"`
	IsMarginEnabled bool     `json:"isMarginEnabled"`
	IsDebitEnabled  bool     `json:"isDebitEnabled"`
	Chains          []*Chain `json:"chains"`
}

type Address struct {
	Address         string `json:"address"`
	Memo            string `json:"memo"`
	Chain           string `json:"chain"`
	ContractAddress string `json:"contract_address"`
}

type Deposit struct {
	Address    string `json:"address"`
	Memo       string `json:"memo"`
	Amount     string `json:"amount"`
	Fee        string `json:"fee"`
	Currency   string `json:"currency"`
	IsInner    bool   `json:"isInner"`
	WalletTxID string `json:"walletTxId"`
	Status     string `json:"status"`
	Remark     string `json:"remark"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

type CreateOrder struct {
	OrderID string `json:"orderId"`
}

type Order struct {
	Id          string `json:"id"`
	Symbol      string `json:"symbol"`
	OpType      string `json:"opType"`
	Type        string `json:"type"`
	Side        string `json:"side"`
	Price       string `json:"price"`
	Size        string `json:"size"`
	Funds       string `json:"funds"`
	DealFunds   string `json:"dealFunds"`
	DealSize    string `json:"dealSize"`
	Fee         string `json:"fee"`
	FeeCurrency string `json:"feeCurrency"`
	ClientOid   string `json:"clientOid"`
	Remark      string `json:"remark"`
	Tags        string `json:"tags"`
	IsActive    bool   `json:"isActive"`
	CancelExist bool   `json:"cancelExist"`
	CreatedAt   int64  `json:"createdAt"`
	TradeType   string `json:"tradeType"`
}

type CreateWithdrawal struct {
	WithdrawalID string `json:"withdrawalId"`
}

type WithdrawalQuota struct {
	Currency            string `json:"currency"`
	AvailableAmount     string `json:"availableAmount"`
	RemainAmount        string `json:"remainAmount"`
	WithdrawMinSize     string `json:"withdrawMinSize"`
	LimitBTCAmount      string `json:"limitBTCAmount"`
	InnerWithdrawMinFee string `json:"innerWithdrawMinFee"`
	UsedBTCAmount       string `json:"usedBTCAmount"`
	IsWithdrawEnabled   bool   `json:"isWithdrawEnabled"`
	WithdrawMinFee      string `json:"withdrawMinFee"`
	Precision           uint8  `json:"precision"`
	Chain               string `json:"chain"`
}

type Withdrawal struct {
	ID         string `json:"id"`
	Address    string `json:"address"`
	Memo       string `json:"memo"`
	Currency   string `json:"currency"`
	Amount     string `json:"amount"`
	Fee        string `json:"fee"`
	WalletTxId string `json:"walletTxId"`
	IsInner    bool   `json:"isInner"`
	Status     string `json:"status"`
	Remark     string `json:"remark"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}
