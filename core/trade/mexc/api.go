package mexc

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"

	"github.com/magicpool-co/pool/core/trade/common"
	coreCommon "github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

const (
	digitsMode  = common.TickSize
	paddingMode = common.NoPadding
)

func precisionTruncate(value, precison string) (string, error) {
	return common.DecimalToPrecision(value, precison, common.Truncate, digitsMode, paddingMode)
}

func precisionRound(value, precison string) (string, error) {
	return common.DecimalToPrecision(value, precison, common.Round, digitsMode, paddingMode)
}

func precisionSplitPercentage(value string, numerator, denominator uint64) (string, error) {
	units := new(big.Int).SetUint64(1e18)
	bigValue, err := coreCommon.StringDecimalToBigint(value, units)
	if err != nil {
		return "", err
	}

	splitValue := coreCommon.SplitBigPercentage(bigValue, numerator, denominator)
	parsedValue := coreCommon.BigIntToFloat64(splitValue, units)
	strValue := strconv.FormatFloat(parsedValue, 'f', 8, 64)

	return strValue, nil
}

func parseSymbolCommission(symbol *Symbol) (uint64, uint64, error) {
	var units uint64 = 1e4

	unitsBig := new(big.Int).SetUint64(units)
	maker, err := coreCommon.StringDecimalToBigint(symbol.MakerCommission, unitsBig)
	if err != nil {
		return 0, 0, err
	}

	taker, err := coreCommon.StringDecimalToBigint(symbol.TakerCommission, unitsBig)
	if err != nil {
		return 0, 0, err
	}

	commission := taker.Uint64()
	if maker.Cmp(taker) > 0 {
		commission = maker.Uint64()
	}

	return commission, units, nil
}

/* general */

func (c *Client) ID() types.ExchangeID {
	return types.MEXCGlobalID
}

func (c *Client) GetTradeTimeout() time.Duration {
	return time.Minute * 2
}

/* account */

func (c *Client) GetAccountStatus() error {
	var obj *Account
	err := c.do("GET", "/api/v3/account", nil, &obj, false, true)
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

func (c *Client) GetRate(market string) (float64, error) {
	payload := map[string]string{
		"symbol": market,
	}

	var obj *Rate
	err := c.do("GET", "/api/v3/avgPrice", payload, &obj, false, false)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(obj.Price, 64)
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
	err := c.do("GET", "/open/api/v2/market/kline", payload, &objs, true, false)
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

func (c *Client) GetWalletStatus(chain string) (bool, bool, error) {
	objs := make([]*Currency, 0)
	err := c.do("GET", "/api/v3/capital/config/getall", nil, &objs, false, true)
	if err != nil {
		return false, false, err
	}

	for _, obj := range objs {
		if formatChain(chain) != obj.Coin {
			continue
		}

		for _, network := range obj.Networks {
			if chain == "USDC" && network.Network == "ERC20" {
				// ok exception
			} else if networkToChain(network.Network) != chain {
				continue
			}

			return network.DepositEnabled, network.WithdrawEnabled, nil
		}
	}

	return false, false, fmt.Errorf("unable to find mainnet chain for %s", chain)
}

func (c *Client) GetWalletBalance(chain string) (float64, float64, error) {
	var obj *Account
	err := c.do("GET", "/api/v3/account", nil, &obj, false, true)
	if err != nil {
		return 0, 0, err
	}

	for _, balance := range obj.Balances {
		if formatChain(chain) != balance.Asset {
			continue
		}

		value, err := strconv.ParseFloat(balance.Free, 64)
		if err != nil {
			return 0, 0, err
		}

		return value, 0, nil
	}

	return 0, 0, nil
}

/* deposit */

func (c *Client) getDepositAddresses(chain string) ([]*Address, error) {
	payload := map[string]string{
		"coin":    formatChain(chain),
		"network": chainToNetwork(chain),
	}

	objs := make([]*Address, 0)
	err := c.do("GET", "/api/v3/capital/deposit/address", payload, &objs, false, true)

	return objs, err
}

func (c *Client) GetDepositAddress(chain string) (string, error) {
	var obj *Address
	addresses, err := c.getDepositAddresses(chain)
	if err != nil {
		return "", err
	} else if len(addresses) > 0 {
		obj = addresses[0]
	} else {
		return "", fmt.Errorf("no deposit address found for %s and unable to create", chain)
	}

	if obj.Coin != formatChain(chain) {
		return "", fmt.Errorf("deposit address chain mismatch: have %s, want %s", obj.Coin, chain)
	} else if obj.Address == "" {
		return "", fmt.Errorf("deposit address empty for chain %s", chain)
	}

	return obj.Address, nil
}

func (c *Client) GetDepositByTxID(chain, txid string) (*types.Deposit, error) {
	payload := map[string]string{
		"coin":  formatChain(chain),
		"limit": "100",
	}

	objs := make([]*Deposit, 0)
	err := c.do("GET", "/api/v3/capital/deposit/hisrec", payload, &objs, false, true)
	if err != nil {
		return nil, err
	}

	for _, deposit := range objs {
		if strings.Split(deposit.TxID, ":")[0] == txid {
			var completed bool
			switch deposit.Status {
			case 1, 2, 3, 4: // pending
				completed = false
			case 5: // success
				completed = true
			case 6: // auditing
				return nil, fmt.Errorf("deposit %s is being audited", txid)
			case 7: // rejected
				return nil, fmt.Errorf("deposit %s was rejected", txid)
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
		"symbol": market,
	}

	var obj *SymbolList
	err := c.do("GET", "/api/v3/exchangeInfo", payload, &obj, false, false)
	if err != nil {
		return nil, err
	}

	for _, symbol := range obj.Symbols {
		if symbol.Symbol == market {
			return symbol, nil
		}
	}

	return nil, fmt.Errorf("market %s not found", market)
}

func (c *Client) GenerateTradePath(fromChain, toChain string) ([]*types.Trade, error) {
	fromChain = formatChain(fromChain)
	toChain = formatChain(toChain)

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
		}

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
			FromChain: unformatChain(localFromChain),
			ToChain:   unformatChain(localToChain),
			Market:    market.Market,
			Direction: market.Direction,
		}
	}

	return trades, nil
}

func (c *Client) getOrderBookTop(market string) (string, string, error) {
	payload := map[string]string{
		"symbol": market,
	}

	var obj *OrderBook
	err := c.do("GET", "/api/v3/ticker/bookTicker", payload, &obj, false, false)
	if err != nil {
		return "", "", err
	}

	return obj.BidPrice, obj.AskPrice, nil
}

func (c *Client) CreateTrade(market string, direction types.TradeDirection, quantity float64) (string, error) {
	payload := map[string]string{
		"symbol":           market,
		"side":             direction.String(),
		"newClientOrderId": strings.ReplaceAll(uuid.NewString(), "-", ""),
	}

	symbol, err := c.getSymbol(market)
	if err != nil {
		return "", err
	} else if !symbol.IsSpotTradingAllowed {
		return "", fmt.Errorf("market %s is not enabled for spot", market)
	}

	strQuantity := strconv.FormatFloat(quantity, 'f', 8, 64)
	basePrecison := fmt.Sprintf("1e-%d", symbol.BaseAssetPrecision)
	quotePrecison := fmt.Sprintf("1e-%d", symbol.QuoteAssetPrecision)

	var marketOrderEnabled bool
	for _, orderType := range symbol.OrderTypes {
		if orderType == "MARKET" && symbol.QuoteOrderQtyMarketAllowed {
			marketOrderEnabled = true
			break
		}
	}

	if marketOrderEnabled {
		payload["type"] = "MARKET"

		switch direction {
		case types.TradeBuy:
			payload["quoteOrderQty"], err = precisionTruncate(strQuantity, quotePrecison)
			if err != nil {
				return "", err
			}
		case types.TradeSell:
			payload["quantity"], err = precisionTruncate(strQuantity, basePrecison)
			if err != nil {
				return "", err
			}
		}
	} else {
		payload["type"] = "LIMIT"

		commission, commissionUnits, err := parseSymbolCommission(symbol)
		if err != nil {
			return "", err
		}

		topBidPrice, topAskPrice, err := c.getOrderBookTop(market)
		if err != nil {
			return "", err
		}

		var slippage float64 = 0.0001
		var price string
		switch direction {
		case types.TradeBuy:
			priceRaw, err := strconv.ParseFloat(topAskPrice, 64)
			if err != nil {
				return "", err
			}
			priceRaw += slippage * priceRaw
			price = strconv.FormatFloat(priceRaw, 'f', 8, 64)

			strQuantity, err = common.PreciseStringDivWithPrecision(strQuantity, price, symbol.BaseAssetPrecision)
			if err != nil {
				return "", err
			}

			strQuantity, err = precisionSplitPercentage(strQuantity, commissionUnits-commission, commissionUnits)
			if err != nil {
				return "", err
			}
		case types.TradeSell:
			priceRaw, err := strconv.ParseFloat(topBidPrice, 64)
			if err != nil {
				return "", err
			}
			priceRaw -= slippage * priceRaw
			price = strconv.FormatFloat(priceRaw, 'f', 8, 64)
		}

		payload["quantity"], err = precisionTruncate(strQuantity, basePrecison)
		if err != nil {
			return "", err
		}

		payload["price"], err = precisionRound(price, quotePrecison)
		if err != nil {
			return "", err
		}
	}

	var obj *Order
	err = c.do("POST", "/api/v3/order", payload, &obj, false, true)
	if err != nil {
		return "", err
	} else if obj.ClientOrderID == "" {
		return "", fmt.Errorf("empty order id")
	}

	return obj.ClientOrderID, nil
}

func (c *Client) getTrade(market, orderID string) (*Trade, error) {
	payload := map[string]string{
		"symbol":  market,
		"orderId": orderID,
	}

	objs := make([]*Trade, 0)
	err := c.do("GET", "/api/v3/myTrades", payload, &objs, false, true)
	if err != nil {
		return nil, err
	} else if len(objs) == 0 {
		return nil, nil
	}

	sumQuantity := "0"
	sumQuoteQuantity := "0"
	sumCommission := "0"
	var maxTime int64
	for _, obj := range objs {
		if obj.Time > maxTime {
			maxTime = obj.Time
		}

		sumQuantity, err = common.PreciseStringAdd(sumQuantity, obj.Quantity)
		if err != nil {
			return nil, err
		}

		sumQuoteQuantity, err = common.PreciseStringAdd(sumQuoteQuantity, obj.QuoteQuantity)
		if err != nil {
			return nil, err
		}

		sumCommission, err = common.PreciseStringAdd(sumCommission, obj.Commission)
		if err != nil {
			return nil, err
		}
	}

	obj := &Trade{
		Symbol:          market,
		OrderID:         orderID,
		OrderListID:     objs[0].OrderListID,
		Quantity:        sumQuantity,
		QuoteQuantity:   sumQuoteQuantity,
		Commission:      sumCommission,
		CommissionAsset: objs[0].CommissionAsset,
		Time:            maxTime,
		ClientOrderID:   objs[0].ClientOrderID,
	}

	return obj, nil
}

func (c *Client) GetTradeByID(market, tradeID string, inputValue float64) (*types.Trade, error) {
	payload := map[string]string{
		"symbol":            market,
		"origClientOrderId": tradeID,
	}

	var obj *Order
	err := c.do("GET", "/api/v3/order", payload, &obj, false, true)
	if err != nil {
		return nil, err
	}

	var completed, active bool
	switch obj.Status {
	case "NEW":
		completed, active = false, true
	case "PARTIALLY_FILLED":
		completed, active = false, true
	case "FILLED":
		completed, active = true, false
	case "CANCELED", "CANCELLED", "PARTIALLY_CANCELED":
		completed, active = false, false
	default:
		return nil, fmt.Errorf("order has an unknown status %s", obj.Status)
	}

	symbol, err := c.getSymbol(obj.Symbol)
	if err != nil {
		return nil, err
	}

	avgFillPrice := "0.0"
	trade, err := c.getTrade(market, obj.OrderID)
	if err != nil {
		return nil, err
	} else if trade != nil {
		if trade.Quantity != obj.ExecutedQty || trade.QuoteQuantity != obj.CumulativeQuoteQty {
			return nil, fmt.Errorf("trade does not match")
		}

		avgFillPrice, err = common.PreciseStringDiv(trade.QuoteQuantity, trade.Quantity)
		if err != nil {
			return nil, err
		}
	}

	var fromChain, toChain string
	var direction types.TradeDirection
	outputValue, proceeds, fee := "0.0", "0.0", "0.0"
	switch strings.ToUpper(obj.Side) {
	case "BUY":
		fromChain, toChain = symbol.QuoteAsset, symbol.BaseAsset
		direction = types.TradeBuy

		if trade != nil {
			outputValue = trade.QuoteQuantity
			proceeds = trade.Quantity
			fee, err = common.PreciseStringDiv(trade.Commission, avgFillPrice)
			if err != nil {
				return nil, err
			}
		}
	case "SELL":
		fromChain, toChain = symbol.BaseAsset, symbol.QuoteAsset
		direction = types.TradeSell

		if trade != nil {
			outputValue = trade.Quantity
			proceeds = trade.QuoteQuantity
			fee = trade.Commission
		}
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
		Fees:     fee,
		Price:    avgFillPrice,

		Completed: completed,
		Active:    active,
	}

	return parsedTrade, nil
}

func (c *Client) CancelTradeByID(market, tradeID string) error {
	payload := map[string]string{
		"symbol":            market,
		"origClientOrderId": tradeID,
	}

	var obj map[string]interface{}
	err := c.do("DELETE", "/api/v3/order", payload, &obj, false, true)

	return err
}

/* withdrawal */

func (c *Client) CreateWithdrawal(chain, address string, quantity float64) (string, error) {
	payload := map[string]string{
		"coin":    chain,
		"network": chainToNetwork(chain),
		"address": address,
		"amount":  strconv.FormatFloat(quantity, 'f', 8, 64),
	}

	var obj *Withdrawal
	err := c.do("POST", "/api/v3/capital/withdraw/apply", payload, &obj, false, true)
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
	err := c.do("GET", "/api/v3/capital/withdraw/history", payload, &objs, false, true)
	if err != nil {
		return nil, err
	}

	for _, withdrawal := range objs {
		if withdrawal.ID == withdrawalID {
			var completed bool
			switch withdrawal.Status {
			case 1, 2, 3, 4, 5, 6: // apply, auditing, wait, processing, wait_packaging, wait_confirm
				completed = false
			case 7: // success
				completed = true
			case 8: // failed
				return nil, fmt.Errorf("withdrawal has failed")
			case 9: // cancel
				return nil, fmt.Errorf("withdrawal was cancelled")
			case 10: // manual
				return nil, fmt.Errorf("withdrawal requires manual intervention")
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
