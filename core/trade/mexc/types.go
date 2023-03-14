package mexc

import (
	"github.com/goccy/go-json"
)

type Response struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"msg"`
}

type Account struct {
	MakerCommission  int    `json:"makerCommission"`
	TakerCommission  int    `json:"takerCommission"`
	BuyerCommission  int    `json:"buyerCommission"`
	SellerCommission int    `json:"sellerCommission"`
	CanDeposit       bool   `json:"canDeposit"`
	CanTrade         bool   `json:"canTrade"`
	CanWithdraw      bool   `json:"canWithdraw"`
	UpdateTime       *int   `json:"updateTime"`
	AccountType      string `json:"accountType"`
	Balances         []struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	} `json:"balances"`
	Permissions []string `json:"permissions"`
}

type Rate struct {
	Mins  int    `json:"mins"`
	Price string `json:"price"`
}

type Network struct {
	Coin             string  `json:"coin"`
	Name             string  `json:"name"`
	Network          string  `json:"network"`
	MinConfirmations int     `json:"minConfirm"`
	DepositEnabled   bool    `json:"depositEnable"`
	WithdrawEnabled  bool    `json:"withdrawEnable"`
	WithdrawFee      string  `json:"withdrawFee"`
	WithdrawMin      string  `json:"withdrawMin"`
	WithdrawMax      string  `json:"withdrawMax"`
	SameAddress      bool    `json:"sameAddress"`
	Contract         *string `json:"contract"`
}

type Currency struct {
	Coin     string     `json:"coin"`
	Name     string     `json:"string"`
	Networks []*Network `json:"networkList"`
}

type Address struct {
	Coin    string `json:"coin"`
	Network string `json:"network"`
	Address string `json:"address"`
	Memo    string `json:"memo"`
}

type Deposit struct {
	Amount        string `json:"amount"`
	Coin          string `json:"coin"`
	Network       string `json:"network"`
	Status        int    `json:"status"`
	Address       string `json"address"`
	AddressTag    string `json:"addressTag"`
	TxID          string `json:"txId"`
	InsertTime    int    `json:"insertTime"`
	TransferType  int    `json:"transferType"`
	UnlockConfirm string `json:"unlockConfirm"`
	ConfirmTimes  string `json:"confirmTimes"`
}

type Symbol struct {
	Symbol                     string   `json:"symbol"`
	Status                     string   `json:"status"`
	BaseAsset                  string   `json:"baseAsset"`
	BaseAssetPrecision         int      `json:"baseAssetPrecision"`
	BaseCommissionPrecision    int      `json:"baseCommissionPrecision"`
	QuoteAsset                 string   `json:"quoteAsset"`
	QuoteAssetPrecision        int      `json:"quoteAssetPrecision"`
	QuoteCommissionPrecision   int      `json:"quoteCommissionPrecision"`
	OrderTypes                 []string `json:"orderTypes"`
	QuoteOrderQtyMarketAllowed bool     `json:"quoteOrderQtyMarketAllowed"`
	AllowTrailingStop          bool     `json:"allowTrailingStop"`
	CancelReplaceAllowed       bool     `json:"cancelReplaceAllowed"`
	IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed     bool     `json:"isMarginTradingAllowed"`
	Permissions                []string `json:"permissions"`
	MakerCommission            string   `json:"makerCommission"`
	TakerCommission            string   `json:"takerCommission"`
}

type SymbolList struct {
	Timezone   string `json:"timezone"`
	ServerTime int64  `json:"serverTime"`
	Symbols    []*Symbol
}

type OrderBook struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	BidQty   string `json:"bidQty"`
	AskPrice string `json:"askPrice"`
	AskQty   string `json:"askQty"`
}

type Order struct {
	Symbol             string `json:"symbol"`
	OrderID            string `json:"orderId"`
	OrderListID        int64  `json:"orderListId"`
	ClientOrderID      string `json:"clientOrderId"`
	Price              string `json:"price"`
	OrigQty            string `json:"origQty"`
	ExecutedQty        string `json:"executedQty"`
	CumulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status             string `json:"status"`
	TimeInForce        string `json:"timeInForce"`
	Type               string `json:"type"`
	Side               string `json:"side"`
	StopPrice          string `json:"stopPrice"`
	IcebergQty         string `json:"icebergQty"`
	Time               int    `json:"time"`
	UpdateTime         int    `json:"updateTime"`
	IsWorking          bool   `json:"isWorking"`
	OrigQuoteOrderQty  string `json:"origQuoteOrderQty"`
}

type Trade struct {
	Symbol          string `json:"symbol"`
	ID              string `json:"id"`
	OrderID         string `json:"orderId"`
	OrderListID     int64  `json:"orderListId"`
	Price           string `json:"price"`
	Quantity        string `json:"qty"`
	QuoteQuantity   string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            int64  `json:"time"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
	IsBestMatch     bool   `json:"isBestMatch"`
	IsSelfTrade     bool   `json:"isSelfTrade"`
	ClientOrderID   string `json:"clientOrderId"`
}

type Withdrawal struct {
	Address         string `json:"address"`
	Amount          string `json:"amount"`
	ApplyTime       int64  `json:"applyTime"`
	Coin            string `json:"coin"`
	ID              string `json"id"`
	WithdrawOrderID string `json:"withdrawOrderId"`
	Network         string `json:"network"`
	TransferType    int    `json:"transferType"`
	Status          int    `json:"status"`
	TransactionFee  string `json:"transactionFee"`
	ConfirmNo       int    `json:"confirmNo"`
	Info            string `json:"info"`
	TxID            string `json:"txId"`
}
