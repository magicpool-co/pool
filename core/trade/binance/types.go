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
	ConfirmNo       int    `json:"confirmNo"`
	Info            string `json:"info"`
	TxID            string `json:"txId"`
}
