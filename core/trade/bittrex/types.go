package bittrex

import (
	"time"
)

type RateResponse struct {
	Symbol        string `json:"symbol"`
	LastTradeRate string `json:"lastTradeRate"`
	BidRate       string `json:"bidRate"`
	AskRate       string `json:"askRate"`
}

type HistoricalRateResponse struct {
	StartsAt time.Time `json:"startsAt"`
	Open     string    `json:"open"`
	High     string    `json:"high"`
	Low      string    `json:"low"`
	Close    string    `json:"close"`
}

type Currency struct {
	Symbol                   string        `json:"symbol"`
	Name                     string        `json:"name"`
	CoinType                 string        `json:"coinType"`
	Status                   string        `json:"status"`
	MinConfirmations         int           `json:"minConfirmations"`
	Notice                   string        `json:"notice"`
	TxFee                    string        `json:"txFee"`
	LogoURL                  string        `json:"logoUrl,omitempty"`
	ProhibitedIn             []interface{} `json:"prohibitedIn"`
	BaseAddress              string        `json:"baseAddress,omitempty"`
	AssociatedTermsOfService []interface{} `json:"associatedTermsOfService"`
}

type Address struct {
	Status           string `json:"status"`
	CurrencySymbol   string `json:"currencySymbol"`
	CryptoAddress    string `json:"cryptoAddress"`
	CryptoAddressTag string `json:"cryptoAddressTag"`
}

type Balance struct {
	Currency  string    `json:"currencySymbol"`
	Total     string    `json:"total"`
	Available string    `json:"available"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Deposit struct {
	ID               string `json:"id"`
	CurrencySymbol   string `json:"currencySymbol"`
	Quantity         string `json:"quantity"`
	CryptoAddress    string `json:"cryptoAddress"`
	CryptoAddressTag string `json:"cryptoAddressTag"`
	TxID             string `json:"txId"`
	Confirmations    int32  `json:"confirmations"`
	UpdatedAt        string `json:"updatedAt"`
	CompletedAt      string `json:"completedAt"`
	Status           string `json:"status"`
	Source           string `json:"source"`
}

type Order struct {
	ID            string      `json:"id"`
	MarketSymbol  string      `json:"marketSymbol"`
	Direction     string      `json:"direction"`
	Type          string      `json:"type"`
	Quantity      string      `json:"quantity"`
	Limit         string      `json:"limit"`
	Ceiling       string      `json:"ceiling"`
	TimeInForce   string      `json:"timeInForce"`
	ClientOrderID string      `json:"clientOrderId"`
	FillQuantity  string      `json:"fillQuantity"`
	Commission    string      `json:"commission"`
	Proceeds      string      `json:"proceeds"`
	Status        string      `json:"status"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
	ClosedAt      time.Time   `json:"closedAt"`
	OrderToCancel interface{} `json:"orderToCancel"`
}

type Withdrawal struct {
	ID               string    `json:"id"`
	CurrencySymbol   string    `json:"currencySymbol"`
	Quantity         string    `json:"quantity"`
	CryptoAddress    string    `json:"cryptoAddress"`
	CryptoAddressTag string    `json:"cryptoAddressTag"`
	TxCost           string    `json:"txCost"`
	TxID             string    `json:"txId"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	CompletedAt      time.Time `json:"completedAt"`
}
