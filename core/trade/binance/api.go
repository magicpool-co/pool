package binance

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/magicpool-co/pool/types"
)

/* account */

func (c *Client) GetAccountStatus() error {
	var obj *Account
	err := c.do("GET", "/api/v3/account", nil, &obj, securityTypeSigned)
	if err != nil {
		return err
	} else if !obj.CanDeposit {
		return fmt.Errorf("unable to deposit on binance account")
	} else if !obj.CanTrade {
		return fmt.Errorf("unable to trade on binance account")
	} else if !obj.CanWithdraw {
		return fmt.Errorf("unable to withdraw on binance account")
	}

	return nil
}

/* rate */

func (c *Client) GetRate(base, quote string) (float64, error) {
	payload := map[string]string{
		"symbol": strings.ToUpper(base + quote),
	}

	var obj *Rate
	err := c.do("GET", "/api/v3/ticker/price", payload, &obj, securityTypeNone)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(obj.Price, 64)
}

func (c *Client) GetHistoricalRate(base, quote string, timestamp time.Time) (float64, error) {
	payload := map[string]string{
		"symbol":    strings.ToUpper(base + quote),
		"interval":  "5m",
		"startTime": strconv.FormatInt(timestamp.Add(-1*time.Hour).UnixMilli(), 10),
		"endTime":   strconv.FormatInt(timestamp.Add(time.Hour).UnixMilli(), 10),
	}

	objs := make([][]interface{}, 0)
	err := c.do("GET", "/api/v3/klines", payload, &objs, securityTypeNone)
	if err != nil {
		return 0, err
	} else if len(objs) == 0 {
		return 0, fmt.Errorf("no results found for rate")
	}

	kline := objs[len(objs)-1]
	if len(kline) != 12 {
		return 0, fmt.Errorf("invalid kline of length %d", len(kline))
	}

	rawClosePrice, ok := kline[4].(string)
	if !ok {
		return 0, fmt.Errorf("unable to cast kline[4] %v as string", kline[4])
	}

	return strconv.ParseFloat(rawClosePrice, 64)
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
				localPrice, err := c.GetRate(market.Base, market.Quote)
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
	payload := map[string]string{
		"asset": strings.ToUpper(chain),
	}

	var obj *Asset
	err := c.do("GET", "/sapi/v1/asset/assetDetail", payload, &obj, securityTypeSigned)
	if err != nil {
		return false, err
	} else if !obj.DepositStatus {
		if obj.DepositTip != "" {
			return false, fmt.Errorf("deposits are disabled for %s: %s", chain, obj.DepositTip)
		}
		return false, fmt.Errorf("deposits are disabled for %s", chain)
	} else if !obj.WithdrawStatus {
		return false, fmt.Errorf("withdrawals are disabled for %s", chain)
	}

	return true, nil
}

func (c *Client) GetWalletBalance(chain string) (float64, float64, error) {
	var obj *Account
	err := c.do("GET", "/api/v3/account", nil, &obj, securityTypeSigned)
	if err != nil {
		return 0, 0, err
	}

	for _, balance := range obj.Balances {
		if chain == balance.Asset {
			value, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				return 0, 0, err
			}

			return value, 0, nil
		}
	}

	return 0, 0, nil
}

/* deposit */

func (c *Client) GetDepositAddress(chain string) (string, error) {
	payload := map[string]string{
		"coin": strings.ToUpper(chain),
	}

	var obj *Address
	err := c.do("GET", "/sapi/v1/capital/deposit/address", payload, &obj, securityTypeSigned)
	if err != nil {
		return "", err
	} else if obj.Address == "" {
		return "", fmt.Errorf("deposit address empty for chain %s", chain)
	}

	return obj.Address, nil
}

func (c *Client) GetDepositByTxID(chain, txid string) (*types.Deposit, error) {
	payload := map[string]string{
		"coin":  strings.ToUpper(chain),
		"limit": "100",
		"txid":  txid,
	}

	objs := make([]*Deposit, 0)
	err := c.do("GET", "/sapi/v1/capital/deposit/hisre", payload, &objs, securityTypeSigned)
	if err != nil {
		return nil, err
	}

	for _, deposit := range objs {
		// @TODO: depending on the chain, txid may not actually match
		// like with ETH when they forward it via smart contract
		if deposit.TxID == txid {
			var completed bool
			switch deposit.Status {
			case 0: // pending
				completed = false
			case 1: // success
				completed = true
			case 6: // credited but cannot withdrawal
				completed = false
			default:
				return nil, fmt.Errorf("unknown deposit status %d", deposit.Status)
			}

			parsedDeposit := &types.Deposit{
				ID:        txid,
				TxID:      txid,
				Value:     deposit.Amount,
				Fee:       "0",
				Completed: completed,
			}

			return parsedDeposit, nil
		}
	}

	return nil, fmt.Errorf("deposit not found")
}

func (c *Client) GetDepositByID(chain, depositID string) (*types.Deposit, error) {
	return c.GetDepositByTxID(chain, depositID)
}

/* transfer */

func (c *Client) TransferToMainAccount(chain string, value float64) error {
	return nil
}

func (c *Client) TransferToTradeAccount(chain string, value float64) error {
	return nil
}

/* order */

func (c *Client) getSymbol(market string) (*Symbol, error) {
	payload := map[string]string{
		"symbol": strings.ToUpper(market),
	}

	var obj *ExchangeInformation
	err := c.do("GET", "/api/v3/exchangeInfo", payload, &obj, securityTypeNone)
	if err != nil {
		return nil, err
	}

	for _, symbol := range obj.Symbols {
		if symbol.Symbol == market {
			return symbol, nil
		}
	}

	return nil, fmt.Errorf("symbol not found")
}

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
		symbol, err := c.getSymbol(market.Market)
		if err != nil {
			return nil, err
		} else if !symbol.IsSpotTradingAllowed {
			return nil, fmt.Errorf("market %s is not enabled for spot", market.Market)
		} else if market.Direction == types.TradeBuy && !symbol.QuoteOrderQtyMarketAllowed {
			return nil, fmt.Errorf("market %s is not enabled for quote order quantity", market.Market)
		}

		var localFromChain, localToChain string
		switch market.Direction {
		case types.TradeBuy:
			localFromChain, localToChain = market.Base, market.Quote
		case types.TradeSell:
			localFromChain, localToChain = market.Quote, market.Base
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

func (c *Client) getTradeFees(market string) (string, error) {
	payload := map[string]string{
		"symbol": market,
	}

	var obj *TradeFee
	err := c.do("GET", "/sapi/v1/asset/tradeFee", payload, &obj, securityTypeSigned)
	if err != nil {
		return "", err
	}

	return obj.TakerCommission, nil
}

func (c *Client) CreateTrade(market string, direction types.TradeDirection, quantity float64) (string, error) {
	payload := map[string]string{
		"symbol":           market,
		"side":             direction.String(),
		"type":             "MARKET",
		"newOrderRespType": "RESULT",
	}

	symbol, err := c.getSymbol(market)
	if err != nil {
		return "", err
	} else if !symbol.IsSpotTradingAllowed {
		return "", fmt.Errorf("market %s is not enabled for spot", market)
	} else if direction == types.TradeBuy && !symbol.QuoteOrderQtyMarketAllowed {
		return "", fmt.Errorf("market %s is not enabled for quote order quantity", market)
	}

	switch direction {
	case types.TradeBuy:
		payload["quoteOrderQty"] = strconv.FormatFloat(quantity, 'f', 8, 64)
	case types.TradeSell:
		payload["quantity"] = strconv.FormatFloat(quantity, 'f', 8, 64)
	default:
		return "", fmt.Errorf("invalid trade direction %d", direction)
	}

	var obj *Order
	err = c.do("POST", "/api/v3/order", payload, &obj, securityTypeSigned)
	if err != nil {
		return "", err
	} else if obj.ClientOrderID == "" {
		return "", fmt.Errorf("empty order id")
	}

	return obj.ClientOrderID, nil
}

func (c *Client) GetTradeByID(market, tradeID string, inputValue float64) (*types.Trade, error) {
	payload := map[string]string{
		"symbol":            market,
		"origClientOrderId": tradeID,
	}

	var obj *Order
	err := c.do("GET", "/api/v3/order", payload, &obj, securityTypeSigned)
	if err != nil {
		return nil, err
	}

	var completed bool
	switch obj.Status {
	case "NEW", "PARTIALLY_FILLED":
		completed = false
	case "FILLED":
		completed = true
	case "PENDING_CANCEL", "CANCELLED":
		return nil, fmt.Errorf("order was cancelled")
	case "REJECTED", "EXPIRED":
		return nil, fmt.Errorf("order %s", strings.ToLower(obj.Status))
	default:
		return nil, fmt.Errorf("order has an unknown status %s", obj.Status)
	}

	symbol, err := c.getSymbol(obj.Symbol)
	if err != nil {
		return nil, err
	}

	baseInitialQuantity, err := strconv.ParseFloat(obj.OrigQty, 64)
	if err != nil {
		return nil, err
	}

	baseFinalQuantity, err := strconv.ParseFloat(obj.ExecutedQty, 64)
	if err != nil {
		return nil, err
	}

	quoteInitialQuantity, err := strconv.ParseFloat(obj.OrigQuoteOrderQty, 64)
	if err != nil {
		return nil, err
	}

	quoteFinalQuantity, err := strconv.ParseFloat(obj.CumulativeQuoteQty, 64)
	if err != nil {
		return nil, err
	}

	var avgFillPrice float64
	if baseFinalQuantity > 0 {
		avgFillPrice = quoteFinalQuantity / baseFinalQuantity
	}

	var fromChain, toChain string
	var direction types.TradeDirection
	var outputValue, proceeds, fees string
	switch strings.ToUpper(obj.Side) {
	case "BUY":
		fromChain, toChain = symbol.QuoteAsset, symbol.BaseAsset
		direction = types.TradeBuy
		outputValue = obj.OrigQuoteOrderQty
		proceeds = obj.ExecutedQty
		fees = strconv.FormatFloat(inputValue-quoteInitialQuantity, 'f', 8, 64)
	case "SELL":
		fromChain, toChain = symbol.BaseAsset, symbol.QuoteAsset
		direction = types.TradeSell
		outputValue = obj.OrigQty
		proceeds = obj.CumulativeQuoteQty
		fees = strconv.FormatFloat(inputValue-baseInitialQuantity, 'f', 8, 64)
	default:
		return nil, fmt.Errorf("unknown trade direction")
	}

	parsedTrade := &types.Trade{
		ID:        tradeID,
		FromChain: fromChain,
		ToChain:   toChain,
		Market:    obj.Symbol,
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
	payload := map[string]string{
		"coin":    chain,
		"address": address,
		"amount":  strconv.FormatFloat(quantity, 'f', 8, 64),
	}

	var obj *Withdrawal
	err := c.do("POST", "/sapi/v1/capital/withdraw/apply", payload, &obj, securityTypeSigned)
	if err != nil {
		return "", err
	} else if obj.ID == "" {
		return "", fmt.Errorf("empty withdrawal id")
	}

	return obj.ID, nil
}

func (c *Client) GetWithdrawalByID(chain, withdrawalID string) (*types.Withdrawal, error) {
	payload := map[string]string{
		"coin":            chain,
		"withdrawOrderId": withdrawalID,
		"limit":           "100",
	}

	objs := make([]*Withdrawal, 0)
	err := c.do("GET", "/sapi/v1/capital/withdraw/history", payload, &objs, securityTypeSigned)
	if err != nil {
		return nil, err
	}

	for _, withdrawal := range objs {
		if withdrawal.ID == withdrawalID {
			var completed bool
			switch withdrawal.Status {
			case 0: // email sent
				return nil, fmt.Errorf("withdrawal is waiting for an email")
			case 1: // cancelled
				return nil, fmt.Errorf("withdrawal was cancelled")
			case 2: // awaiting approval
				return nil, fmt.Errorf("withdrawal is awaiting approval")
			case 3: // rejected
				return nil, fmt.Errorf("withdrawal was rejected")
			case 4: // processing
				completed = false
			case 5: // failure
				return nil, fmt.Errorf("withdrawal has failed")
			case 6: // completed
				completed = true
			default:
				return nil, fmt.Errorf("withdrawal has an unknown status status %d", withdrawal.Status)
			}

			parsedWithdrawal := &types.Withdrawal{
				ID:        withdrawal.ID,
				TxID:      withdrawal.TxID,
				Value:     withdrawal.Amount,
				Fee:       withdrawal.TransactionFee,
				Completed: completed,
			}

			return parsedWithdrawal, nil
		}
	}

	return nil, fmt.Errorf("withdrawal not found")
}
