package chart

import (
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

func getMarketRate(chain string, timestamp time.Time) float64 {
	const baseURL = "https://api.coingecko.com/api/v3"
	var tickers = map[string]string{
		"AE":   "aeternity",
		"BTC":  "bitcoin",
		"CFX":  "conflux-token",
		"CTXC": "cortex",
		"ETC":  "ethereum-classic",
		"ETH":  "ethereum",
		"FIRO": "zcoin",
		"FLUX": "flux",
		"RVN":  "ravencoin",
	}

	ticker, ok := tickers[chain]
	if !ok {
		return 0
	}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", baseURL, ticker)
	res, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer res.Body.Close()

	data := make(map[string]map[string]float64)
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return 0
	} else if _, ok := data[ticker]; !ok {
		return 0
	}

	return data[ticker]["usd"]
}
