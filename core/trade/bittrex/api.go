package bittrex

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

/* general */

func (c *Client) ID() types.ExchangeID {
	return types.BittrexID
}

/* account */

func (c *Client) GetAccountStatus() error {
	return nil
}

/* rate */

func (c *Client) GetRate(market string) (float64, error) {
	var obj *RateResponse
	err := c.do("GET", "/markets/"+market+"/ticker", nil, &obj, false)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(obj.LastTradeRate, 64)
}

func (c *Client) GetHistoricalRate(base, quote string, timestamp time.Time) (float64, error) {
	path := "/markets/" + base + "-" + quote + "/candles/MINUTE_5"
	if time.Since(timestamp) < time.Hour*24 {
		path += "/recent"
	} else {
		path += fmt.Sprintf("/historical/%d/%d/%d", timestamp.Year(), timestamp.Month(), timestamp.Day())
	}

	objs := make([]*HistoricalRateResponse, 0)
	err := c.do("GET", path, nil, &objs, false)
	if err != nil {
		return 0, err
	} else if len(objs) == 0 {
		return 0, fmt.Errorf("no results found for rate")
	}

	closestRate := objs[len(objs)-1]

	return strconv.ParseFloat(closestRate.Close, 64)
}

func (c *Client) GetOutputThresholds() map[string]*big.Int {
	return presetOutputThresholds
}

func (c *Client) GetPrices(inputPaths map[string]map[string]*big.Int) (map[string]map[string]float64, error) {
	prices := make(map[string]map[string]float64)

	for fromChain, outputPaths := range inputPaths {
		prices[fromChain] = make(map[string]float64)
		for toChain := range outputPaths {
			if _, ok := presetTradePaths[fromChain]; !ok {
				return nil, fmt.Errorf("no trade path found for %s->%s", fromChain, toChain)
			}

			markets := presetTradePaths[fromChain][toChain]
			if len(markets) == 0 {
				return nil, fmt.Errorf("no trade path found for %s->%s", fromChain, toChain)
			}

			prices[fromChain][toChain] = 1
			for _, market := range markets {
				localPrice, err := c.GetRate(market.Market)
				if err != nil {
					return nil, err
				}

				switch market.Direction {
				case types.TradeBuy:
					prices[fromChain][toChain] /= localPrice
				case types.TradeSell:
					prices[fromChain][toChain] *= localPrice
				}
			}
		}
	}

	return prices, nil
}

/* wallet */

func (c *Client) GetWalletStatus(chain string) (bool, error) {
	var obj *Currency
	err := c.do("GET", "/currencies/"+chain, nil, &obj, false)
	if err != nil {
		return false, err
	}

	return obj.Status == "ONLINE", nil
}

func (c *Client) GetWalletBalance(chain string) (float64, float64, error) {
	var obj *Balance
	err := c.do("GET", "/balances/"+chain, nil, &obj, true)
	if err != nil {
		return 0, 0, err
	}

	balance, err := strconv.ParseFloat(obj.Available, 64)
	if err != nil {
		return 0, 0, err
	}

	return balance, 0, nil
}

func (c *Client) GetDepositAddress(chain string) (string, error) {
	var obj *Address
	err := c.do("GET", "/addresses/"+chain, nil, &obj, true)
	if err != nil && err != ErrEmptyTarget {
		return "", err
	}

	// create the address if it doesn't exist yet
	if obj == nil || obj.CryptoAddress == "" {
		payload := map[string]interface{}{
			"currencySymbol": chain,
		}

		err := c.do("POST", "/addresses", payload, &obj, true)
		if err != nil {
			return "", err
		}
	}

	if obj.CryptoAddress == "" {
		return "", fmt.Errorf("deposit address empty for chain %s", chain)
	}

	return obj.CryptoAddress, nil
}

/* deposit */

func (c *Client) parseDeposit(deposit *Deposit) (*types.Deposit, error) {
	var completed bool
	switch deposit.Status {
	case "PENDING":
		completed = false
	case "COMPLETED":
		completed = true
	case "ORPHANED", "INVALIDATED":
		return nil, fmt.Errorf("deposit was %s", deposit.Status)
	default:
		return nil, fmt.Errorf("unknown deposit status %s", deposit.Status)
	}

	parsedDeposit := &types.Deposit{
		ID:        deposit.ID,
		TxID:      deposit.TxID,
		Value:     deposit.Quantity,
		Fee:       "0",
		Completed: completed,
	}

	return parsedDeposit, nil
}

func (c *Client) GetDepositByTxID(chain, txid string) (*types.Deposit, error) {
	objs := make([]*Deposit, 0)
	err := c.do("GET", "/deposits/ByTxId/"+txid, nil, &objs, true)
	if err != nil {
		return nil, err
	} else if len(objs) == 0 {
		return nil, fmt.Errorf("deposit not found")
	} else if len(objs) > 1 {
		return nil, fmt.Errorf("more than 1 deposit found for txid %s", txid)
	}

	return c.parseDeposit(objs[0])
}

func (c *Client) GetDepositByID(chain, depositID string) (*types.Deposit, error) {
	var obj *Deposit
	err := c.do("GET", "/deposits/"+depositID, nil, &obj, true)
	if err != nil {
		return nil, err
	}

	return c.parseDeposit(obj)
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
	fromChain = strings.ToUpper(fromChain)
	toChain = strings.ToUpper(toChain)

	if _, ok := presetTradePaths[fromChain]; !ok {
		return nil, fmt.Errorf("no trade path found for %s->%s", fromChain, toChain)
	}

	markets := presetTradePaths[fromChain][toChain]
	if len(markets) == 0 {
		return nil, fmt.Errorf("no trade path found for %s->%s", fromChain, toChain)
	}

	trades := make([]*types.Trade, len(markets))
	for i, market := range markets {
		var localFromChain, localToChain string
		switch market.Direction {
		case types.TradeBuy:
			localFromChain, localToChain = market.Quote, market.Base
		case types.TradeSell:
			localFromChain, localToChain = market.Base, market.Quote
		default:
			return nil, fmt.Errorf("unknown trade direction %d", market.Direction)
		}

		trades[i] = &types.Trade{
			FromChain: localFromChain,
			ToChain:   localToChain,
			Market:    market.Market,
			Direction: market.Direction,
		}
	}

	return trades, nil
}

func (c *Client) CreateTrade(market string, direction types.TradeDirection, quantity float64) (string, error) {
	payload := map[string]interface{}{
		"marketSymbol": market,
		"direction":    "BUY",
		"timeInForce":  "FILL_OR_KILL",
	}

	switch direction {
	case types.TradeBuy:
		payload["type"] = "CEILING_MARKET"
		payload["ceiling"] = quantity
	case types.TradeSell:
		payload["type"] = "MARKET"
		payload["quantity"] = quantity
	default:
		return "", fmt.Errorf("invalid trade direction %d", direction)
	}

	var obj *Order
	err := c.do("POST", "/orders/", payload, &obj, true)
	if err != nil {
		return "", err
	}

	return obj.ID, nil
}

func (c *Client) GetTradeByID(market, tradeID string, inputValue float64) (*types.Trade, error) {
	var obj *Order
	err := c.do("GET", "/orders/"+tradeID, nil, &obj, true)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(obj.MarketSymbol, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid market %s", obj.MarketSymbol)
	}
	base, quote := parts[0], parts[1]

	proceedsFloat, err := strconv.ParseFloat(obj.Proceeds, 64)
	if err != nil {
		return nil, err
	}

	fillQtyFloat, err := strconv.ParseFloat(obj.FillQuantity, 64)
	if err != nil {
		return nil, err
	}

	var avgFillPrice float64
	if fillQtyFloat > 0 {
		avgFillPrice = proceedsFloat / fillQtyFloat
	}

	var completed bool
	switch obj.Status {
	case "OPEN":
		completed = false
	case "CLOSED":
		completed = true
	default:
		return nil, fmt.Errorf("order has an unknown status %s", obj.Status)
	}

	var fromChain, toChain string
	var direction types.TradeDirection
	var outputValue, proceeds, fees string
	switch strings.ToUpper(obj.Direction) {
	case "BUY":
		var feesFloat float64
		if avgFillPrice > 0 {
			commissionFloat, err := strconv.ParseFloat(obj.Commission, 64)
			if err != nil {
				return nil, err
			}

			feesFloat = commissionFloat / avgFillPrice
		}

		fromChain, toChain = quote, base
		direction = types.TradeBuy
		outputValue = obj.Ceiling
		proceeds = obj.FillQuantity
		fees = strconv.FormatFloat(feesFloat, 'f', 8, 64)
	case "SELL":
		quoteUnits, err := common.GetDefaultUnits(quote)
		if err != nil {
			return nil, err
		}

		proceedsBig, err := common.StringDecimalToBigint(obj.Proceeds, quoteUnits)
		if err != nil {
			return nil, err
		}

		commissionBig, err := common.StringDecimalToBigint(obj.Commission, quoteUnits)
		if err != nil {
			return nil, err
		}

		proceedsBig.Sub(proceedsBig, commissionBig)
		proceedsFloat := common.BigIntToFloat64(proceedsBig, quoteUnits)

		fromChain, toChain = base, quote
		direction = types.TradeSell
		outputValue = obj.Quantity
		proceeds = strconv.FormatFloat(proceedsFloat, 'f', 8, 64)
		fees = obj.Commission
	default:
		return nil, fmt.Errorf("unknown trade direction")
	}

	parsedTrade := &types.Trade{
		ID:        tradeID,
		FromChain: fromChain,
		ToChain:   toChain,
		Market:    obj.MarketSymbol,
		Direction: direction,

		Value:    outputValue,
		Proceeds: proceeds,
		Fees:     fees,
		Price:    strconv.FormatFloat(avgFillPrice, 'f', 8, 64),

		Completed: completed,
	}

	return parsedTrade, nil
}

/* withdrawal */

func (c *Client) CreateWithdrawal(chain, address string, quantity float64) (string, error) {
	payload := map[string]interface{}{
		"currencySymbol": chain,
		"cryptoAddress":  address,
		"quantity":       quantity,
	}

	var obj *Withdrawal
	err := c.do("POST", "/withdrawals", payload, &obj, true)
	if err != nil {
		return "", err
	}

	return obj.ID, nil
}

func (c *Client) GetWithdrawalByID(chain, withdrawalID string) (*types.Withdrawal, error) {
	var obj *Withdrawal
	err := c.do("GET", "/withdrawals/"+withdrawalID, nil, &obj, true)
	if err != nil {
		return nil, err
	}

	var completed bool
	switch obj.Status {
	case "REQUESTED", "AUTHORIZED", "PENDING":
		completed = false
	case "COMPLETED":
		completed = true
	case "CANCELLED":
		return nil, fmt.Errorf("withdrawal cancelled")
	case "ERROR_INVALID_ADDRESS":
		return nil, fmt.Errorf("withdrawal has an invalid address")
	default:
		return nil, fmt.Errorf("unknown withdrawal status %s", obj.Status)
	}

	parsedWithdrawal := &types.Withdrawal{
		ID:        obj.ID,
		TxID:      obj.TxID,
		Value:     obj.Quantity,
		Fee:       obj.TxCost,
		Completed: completed,
	}

	return parsedWithdrawal, nil
}
