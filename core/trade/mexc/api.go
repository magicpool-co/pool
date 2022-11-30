package mexc

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/types"
)

/* general */

func (c *Client) ID() types.ExchangeID {
	return types.MEXCGlobalID
}

/* account */

func (c *Client) GetAccountStatus() error {
	return nil
}

/* rate */

func (c *Client) GetRate(market string) (float64, error) {
	payload := map[string]string{
		"symbol": market,
	}

	var objs []*Symbol
	err := c.do("GET", "/open/api/v2/market/ticker", payload, &objs, false)
	if err != nil {
		return 0, err
	} else if len(objs) != 1 {
		return 0, fmt.Errorf("symbol not found")
	}

	return strconv.ParseFloat(objs[0].Last, 64)
}

func (c *Client) GetHistoricalRates(market string, startTime, endTime time.Time, invert bool) (map[time.Time]float64, error) {
	const maxResults = 1000

	diff := endTime.Sub(startTime)
	if diff/time.Minute*15 > maxResults {
		endTime = startTime.Add(time.Minute * 15 * maxResults)
	}

	payload := map[string]string{
		"symbol":     market,
		"interval":   "15m",
		"start_time": strconv.FormatInt(startTime.Unix(), 10),
		"limit":      strconv.FormatInt(maxResults, 10),
	}

	objs := make([][]json.RawMessage, 0)
	err := c.do("GET", "/open/api/v2/market/kline", payload, &objs, false)
	if err != nil {
		return nil, err
	}

	rates := make(map[time.Time]float64, len(objs))
	for _, kline := range objs {
		if len(kline) != 7 {
			return nil, fmt.Errorf("invalid kline of length %d", len(kline))
		}

		var rawTimestamp int64
		if err := json.Unmarshal(kline[0], &rawTimestamp); err != nil {
			return nil, err
		}

		var rawRate string
		if err := json.Unmarshal(kline[4], &rawRate); err != nil {
			return nil, err
		}

		timestamp := time.Unix(rawTimestamp, 0)
		rate, err := strconv.ParseFloat(rawRate, 64)
		if err != nil {
			return nil, err
		} else if invert && rate > 0 {
			rate = 1 / rate
		}

		rates[timestamp] = rate
	}

	return rates, nil
}

func (c *Client) GetOutputThresholds() map[string]*big.Int {
	return nil
}

func (c *Client) GetPrices(inputPaths map[string]map[string]*big.Int) (map[string]map[string]float64, error) {
	return nil, nil
}

/* wallet */

func (c *Client) GetWalletStatus(chain string) (bool, bool, error) {
	return false, false, nil
}

func (c *Client) GetWalletBalance(chain string) (float64, float64, error) {
	return 0, 0, nil
}

/* deposit */

func (c *Client) GetDepositAddress(chain string) (string, error) {
	return "", nil
}

func (c *Client) GetDepositByTxID(chain, txid string) (*types.Deposit, error) {
	return nil, nil
}

func (c *Client) GetDepositByID(chain, depositID string) (*types.Deposit, error) {
	return nil, nil
}

/* transfer */

func (c *Client) TransferToMainAccount(chain string, value float64) error {
	return nil
}

func (c *Client) TransferToTradeAccount(chain string, value float64) error {
	return nil
}

/* order */

func (c *Client) GenerateTradePath(fromChain, toChain string) ([]*types.Trade, error) {
	return nil, nil
}

func (c *Client) CreateTrade(market string, direction types.TradeDirection, quantity float64) (string, error) {
	return "", nil
}

func (c *Client) GetTradeByID(market, tradeID string, inputValue float64) (*types.Trade, error) {
	return nil, nil
}

/* withdrawal */

func (c *Client) CreateWithdrawal(chain, address string, quantity float64) (string, error) {
	return "", nil
}

func (c *Client) GetWithdrawalByID(chain, withdrawalID string) (*types.Withdrawal, error) {
	return nil, nil
}
