package binance

type Account struct {
	MakerCommission  int    `json:"makerCommission"`
	TakerCommission  int    `json:"takerCommission"`
	BuyerCommission  int    `json:"buyerCommission"`
	SellerCommission int    `json:"sellerCommission"`
	CanDeposit       bool   `json:"canDeposit"`
	CanTrade         bool   `json:"canTrade"`
	CanWithdraw      bool   `json:"canWithdraw"`
	UpdateTime       int    `json:"updateTime"`
	AccountType      string `json:"accountType"`
	Balances         []struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	} `json:"balances"`
	Permissions []string `json:"permissions"`
}

type Rate struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type Asset struct {
	MinWithdrawAmount string `json:"minWithdrawAmount"`
	DepositStatus     bool   `json:"depositStatus"`
	WithdrawFee       int    `json:"withdrawFee"`
	WithdrawStatus    bool   `json:"withdrawStatus"`
	DepositTip        string `json:"depositTip"`
}

type Address struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
	Tag     string `json:"tag"`
	URL     string `json:"url"`
}

type ExchangeInformation struct {
	Timezone   string    `json:"timezone"`
	ServerTime int64     `json:"serverTime"`
	Symbols    []*Symbol `json:"symbols"`
}

type Symbol struct {
	Symbol                     string   `json:"symbol"`
	Status                     string   `json:"status"`
	BaseAsset                  string   `json:"baseAsset"`
	BaseAssetPrecision         int      `json:"baseAssetPrecision"`
	QuoteAsset                 string   `json:"quoteAsset"`
	QuotePrecision             string   `json:"quotePrecision"`
	QuoteAssetPrecision        string   `json:"quoteAssetPrecision"`
	QuoteCommissionPrecision   string   `json:"quoteCommissionPrecision"`
	OrderTypes                 []string `json:"orderTypes"`
	IcebergAllowed             bool     `json:"icebergAllowed"`
	OcoAllowed                 bool     `json:"ocoAllowed"`
	QuoteOrderQtyMarketAllowed bool     `json:"quoteOrderQtyMarketAllowed"`
	AllowTrailingStop          bool     `json:"allowTrailingStop"`
	CancelReplaceAllowed       bool     `json:"cancelReplaceAllowed"`
	IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed     bool     `json:"isMarginTradingAllowed"`
}

type Deposit struct {
	Amount        string `json:"amount"`
	Coin          string `json:"coin"`
	Network       string `json:"network"`
	Status        int    `json:"status"`
	Address       string `json"address"`
	AddressTag    string `json:"addressTag"`
	TxID          string `json:"txId"`
	InsertTime    string `json:"insertTime"`
	TransferType  int    `json:"transferType"`
	UnlockConfirm string `json:"unlockConfirm"`
	ConfirmTimes  string `json:"confirmTimes"`
}

type TradeFee struct {
	Symbol          string `json:"symbol"`
	MakerCommission string `json:"makerCommission"`
	TakerCommission string `json:"takerCommission"`
}

type Order struct {
	Symbol             string `json:"symbol"`
	OrderID            int64  `json"orderId"`
	OrderListID        int64  `json"orderListId"`
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

type Withdrawal struct {
	Address         string `json:"address"`
	Amount          string `json:"amount"`
	ApplyTime       string `json:"applyTime"`
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
