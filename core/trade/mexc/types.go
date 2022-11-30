package mexc

import (
	"github.com/goccy/go-json"
)

type Response struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"msg"`
}

type Symbol struct {
	Symbol     string `json:"symbol"`
	Volume     string `json:"volume"`
	High       string `json:"high"`
	Low        string `json:"low"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Open       string `json:"open"`
	Last       string `json:"last"`
	Time       int64  `json:"time"`
	ChangeRate string `json:"change_rate"`
}
