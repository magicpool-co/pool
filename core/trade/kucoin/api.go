package kucoin

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

/* general */

func (c *Client) ID() types.ExchangeID {
	return types.KucoinID
}

/* account */

func (c *Client) GetAccountStatus() error {
	var obj *Status
	err := c.do("GET", "/api/v1/status", nil, &obj, false)
	if err != nil {
		return err
	} else if obj.Status != "open" {
		return fmt.Errorf("bad exchange status: %s: %s", obj.Status, obj.Message)
	}

	return nil
}

/* rate */

func (c *Client) GetRate(market string) (float64, error) {
	payload := map[string]string{
		"symbol": market,
	}

	var obj *Symbol
	err := c.do("GET", "/api/v1/market/orderbook/level1", payload, &obj, false)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(obj.Price, 64)
}

func (c *Client) GetHistoricalRate(base, quote string, timestamp time.Time) (float64, error) {
	payload := map[string]string{
		"symbol":  formatChain(base) + "-" + formatChain(quote),
		"type":    "5min",
		"startAt": strconv.FormatInt(timestamp.Add(-1*time.Hour).Unix(), 10),
		"endAt":   strconv.FormatInt(timestamp.Add(time.Hour).Unix(), 10),
	}

	objs := make([][]interface{}, 0)
	err := c.do("GET", "/api/v1/market/candles", payload, &objs, false)
	if err != nil {
		return 0, err
	} else if len(objs) == 0 {
		return 0, fmt.Errorf("no results found for rate")
	}

	kline := objs[len(objs)-1]
	if len(kline) != 7 {
		return 0, fmt.Errorf("invalid kline of length %d", len(kline))
	}

	rawClosePrice, ok := kline[2].(string)
	if !ok {
		return 0, fmt.Errorf("unable to cast kline[2] %v as string", kline[2])
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
			if _, ok := presetTradePaths[formatChain(fromChain)]; !ok {
				return nil, fmt.Errorf("no trade path found for %s->%s", fromChain, toChain)
			}

			markets := presetTradePaths[formatChain(fromChain)][formatChain(toChain)]
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
	err := c.do("GET", "/api/v2/currencies/"+formatChain(chain), nil, &obj, false)
	if err != nil {
		return false, err
	}

	for _, chainObj := range obj.Chains {
		if unformatChain(chainObj.ChainName) == chain {
			if !chainObj.IsDepositEnabled {
				return false, fmt.Errorf("deposits are disabled for %s", chain)
			} else if !chainObj.IsWithdrawEnabled {
				return false, fmt.Errorf("withdrawals are disabled for %s", chain)
			}
			return true, nil
		}
	}

	return false, fmt.Errorf("unable to find mainnet chain for %s", chain)
}

func (c *Client) GetWalletBalance(chain string) (float64, float64, error) {
	payload := map[string]string{
		"currency": formatChain(chain),
		"network":  chainToNetwork(chain),
	}

	objs := make([]*Balance, 0)
	err := c.do("GET", "/api/v1/accounts", payload, &objs, true)
	if err != nil {
		return 0, 0, err
	} else if len(objs) == 0 {
		return 0, 0, nil
	}

	var mainBalance, tradeBalance float64
	for _, obj := range objs {
		switch obj.Type {
		case "main":
			mainBalance, err = strconv.ParseFloat(obj.Balance, 64)
			if err != nil {
				return 0, 0, err
			}
		case "trade":
			tradeBalance, err = strconv.ParseFloat(obj.Balance, 64)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	return mainBalance, tradeBalance, nil
}

/* deposit */

func (c *Client) getDepositAddresses(chain string) ([]*Address, error) {
	payload := map[string]string{
		"currency": formatChain(chain),
	}

	objs := make([]*Address, 0)
	err := c.do("GET", "/api/v2/deposit-addresses", payload, &objs, true)

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
		payload := map[string]string{
			"currency": formatChain(chain),
		}

		err = c.do("POST", "/api/v1/deposit-addresses", payload, &obj, true)
		if err != nil {
			return "", err
		}
	}

	if obj.Chain != chain {
		return "", fmt.Errorf("deposit address chain mismatch: have %s, want %s", addresses[0].Chain, chain)
	} else if obj.Address == "" {
		return "", fmt.Errorf("deposit address empty for chain %s", chain)
	}

	return obj.Address, nil
}

func (c *Client) getDeposits(chain string) ([]*Deposit, error) {
	payload := map[string]string{
		"currency": formatChain(chain),
	}

	var paginated *PaginatedResponse
	err := c.do("GET", "/api/v1/deposits", payload, &paginated, true)
	if err != nil {
		return nil, err
	}

	objs := make([]*Deposit, 0)
	err = json.Unmarshal(paginated.Items, &objs)

	return objs, err
}

func (c *Client) GetDepositByTxID(chain, txid string) (*types.Deposit, error) {
	deposits, err := c.getDeposits(chain)
	if err != nil {
		return nil, err
	}

	for _, deposit := range deposits {
		if strings.Split(deposit.WalletTxID, "@")[0] == txid {
			var completed bool
			switch deposit.Status {
			case "SUCCESS":
				completed = true
			case "PROCESSING":
				completed = false
			case "FAILURE":
				return nil, fmt.Errorf("deposit failed")
			default:
				return nil, fmt.Errorf("unknown deposit status %s", deposit.Status)
			}

			parsedDeposit := &types.Deposit{
				ID:        txid,
				TxID:      txid,
				Value:     deposit.Amount,
				Fee:       deposit.Fee,
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

func (c *Client) transferToAccount(chain, from, to string, value float64) error {
	payload := map[string]string{
		"clientOid": uuid.NewString(),
		"currency":  formatChain(chain),
		"from":      from,
		"to":        to,
		"amount":    strconv.FormatFloat(value, 'f', 8, 64),
	}

	var obj *Order
	err := c.do("POST", "/api/v2/accounts/inner-transfer", payload, &obj, true)

	return err
}

func (c *Client) TransferToMainAccount(chain string, value float64) error {
	return c.transferToAccount(chain, "trade", "main", value)
}

func (c *Client) TransferToTradeAccount(chain string, value float64) error {
	return c.transferToAccount(chain, "main", "trade", value)
}

/* order */

func (c *Client) getSymbols() (map[string]*Market, error) {
	var objs []*Market
	err := c.do("GET", "/api/v1/symbols", nil, &objs, false)
	if err != nil {
		return nil, err
	}

	symbols := make(map[string]*Market, len(objs))
	for _, symbol := range objs {
		symbols[symbol.Symbol] = symbol
	}

	return symbols, nil
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

	symbols, err := c.getSymbols()
	if err != nil {
		return nil, err
	}

	trades := make([]*types.Trade, len(markets))
	for i, market := range markets {
		symbol, ok := symbols[market.Market]
		if !ok {
			return nil, fmt.Errorf("market %s not found", market.Market)
		} else if !symbol.EnableTrading {
			return nil, fmt.Errorf("market %s is not enabled", market.Market)
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

func (c *Client) getTradeFees(market string) (string, error) {
	payload := map[string]string{
		"symbols": market,
	}

	objs := make([]*TradeFee, 0)
	err := c.do("GET", "/api/v1/trade-fees", payload, &objs, true)
	if err != nil {
		return "", err
	} else if len(objs) == 0 {
		return "", fmt.Errorf("trade fees not found for market %s", market)
	}

	return objs[0].TakerFee, nil
}

func (c *Client) getAverageFillPrice(tradeID string) (float64, error) {
	payload := map[string]string{
		"orderId": tradeID,
	}

	var paginated *PaginatedResponse
	err := c.do("GET", "/api/v1/fills", payload, &paginated, true)
	if err != nil {
		return 0, err
	}

	objs := make([]*OrderFill, 0)
	err = json.Unmarshal(paginated.Items, &objs)
	if err != nil {
		return 0, err
	}

	var fillPriceWeightedSum float64
	var fillPriceWeightedCount float64
	for _, fill := range objs {
		price, err := strconv.ParseFloat(fill.Price, 64)
		if err != nil {
			return 0, err
		}

		size, err := strconv.ParseFloat(fill.Size, 64)
		if err != nil {
			return 0, err
		}

		fillPriceWeightedSum += price * size
		fillPriceWeightedCount += size
	}

	if fillPriceWeightedCount > 0 {
		return fillPriceWeightedSum / fillPriceWeightedCount, nil
	}

	return 0, nil
}

func (c *Client) CreateTrade(market string, direction types.TradeDirection, quantity float64) (string, error) {
	payload := map[string]string{
		"clientOid": uuid.NewString(),
		"side":      direction.String(),
		"symbol":    market,
		"type":      "market",
	}

	var baseIncrement, quoteIncrement int
	symbols, err := c.getSymbols()
	if err != nil {
		return "", err
	} else if symbol, ok := symbols[market]; !ok {
		return "", fmt.Errorf("market %s not found", market)
	} else if !symbol.EnableTrading {
		return "", fmt.Errorf("market %s is not enabled", market)
	} else {
		baseIncrement, err = parseIncrement(symbol.BaseIncrement)
		if err != nil {
			return "", err
		}

		quoteIncrement, err = parseIncrement(symbol.QuoteIncrement)
		if err != nil {
			return "", err
		}
	}

	switch direction {
	case types.TradeBuy:
		feeRate, err := c.getTradeFees(market)
		if err != nil {
			return "", err
		}

		quote := strings.Split(market, "-")[1]
		quantityStr := strconv.FormatFloat(quantity, 'f', 8, 64)
		newQuantity, _, err := safeSubtractFee(quote, quantityStr, feeRate)
		if err != nil {
			return "", err
		}

		quantity = common.FloorFloatByIncrement(newQuantity, quoteIncrement, 1e8)
		payload["funds"] = strconv.FormatFloat(newQuantity, 'f', 8, 64)
	case types.TradeSell:
		quantity = common.FloorFloatByIncrement(quantity, baseIncrement, 1e8)
		payload["size"] = strconv.FormatFloat(quantity, 'f', 8, 64)
	default:
		return "", fmt.Errorf("invalid trade direction %d", direction)
	}

	var obj *CreateOrder
	err = c.do("POST", "/api/v1/orders", payload, &obj, true)
	if err != nil {
		return "", err
	} else if obj.OrderID == "" {
		return "", fmt.Errorf("empty order id")
	}

	return obj.OrderID, nil
}

func (c *Client) GetTradeByID(market, tradeID string, inputValue float64) (*types.Trade, error) {
	var obj *Order
	err := c.do("GET", "/api/v1/orders/"+tradeID, nil, &obj, true)
	if err != nil {
		return nil, err
	} else if obj.CancelExist {
		return nil, fmt.Errorf("order was cancelled")
	}

	parts := strings.Split(obj.Symbol, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid market %s", obj.Symbol)
	}
	base, quote := parts[0], parts[1]

	avgFillPrice, err := c.getAverageFillPrice(tradeID)
	if err != nil {
		return nil, err
	}

	var fromChain, toChain string
	var direction types.TradeDirection
	var outputValue, proceeds, fees string
	switch strings.ToUpper(obj.Side) {
	case "BUY":
		var feesFloat float64
		if avgFillPrice > 0 {
			quoteInitialQuantity, err := strconv.ParseFloat(obj.Funds, 64)
			if err != nil {
				return nil, err
			}
			feesFloat = (inputValue - quoteInitialQuantity) / avgFillPrice
		}

		fromChain, toChain = quote, base
		direction = types.TradeBuy
		outputValue = obj.Funds
		proceeds = obj.DealSize
		fees = strconv.FormatFloat(feesFloat, 'f', 8, 64)
	case "SELL":
		feeRate, err := c.getTradeFees(obj.Symbol)
		if err != nil {
			return nil, err
		}

		proceedsFloat, feesFloat, err := safeSubtractFee(obj.FeeCurrency, obj.DealFunds, feeRate)
		if err != nil {
			return nil, err
		}

		fromChain, toChain = base, quote
		direction = types.TradeSell
		outputValue = obj.Size
		proceeds = strconv.FormatFloat(proceedsFloat, 'f', 8, 64)
		fees = strconv.FormatFloat(feesFloat, 'f', 8, 64)
	default:
		return nil, fmt.Errorf("unknown trade direction")
	}

	parsedTrade := &types.Trade{
		ID:        obj.ID,
		FromChain: fromChain,
		ToChain:   toChain,
		Market:    obj.Symbol,
		Direction: direction,

		Value:    outputValue,
		Proceeds: proceeds,
		Fees:     fees,
		Price:    strconv.FormatFloat(avgFillPrice, 'f', 8, 64),

		Completed: !obj.IsActive,
	}

	return parsedTrade, nil
}

/* withdrawal */

func (c *Client) checkWithdrawalQuota(chain string, quantity float64) error {
	payload := map[string]string{
		"currency": formatChain(chain),
	}

	var obj *WithdrawalQuota
	err := c.do("GET", "/api/v1/withdrawals/quotas", payload, &obj, true)
	if err != nil {
		return err
	} else if !obj.IsWithdrawEnabled {
		return fmt.Errorf("withdrawals are not enabled for %s", chain)
	} else if remainingQuota, err := strconv.ParseFloat(obj.AvailableAmount, 64); err != nil {
		return err
	} else if remainingQuota < quantity {
		return fmt.Errorf("%f is greater than the withdrawal limit %f for %s", quantity, remainingQuota, chain)
	}

	return nil
}

func (c *Client) CreateWithdrawal(chain, address string, quantity float64) (string, error) {
	err := c.checkWithdrawalQuota(chain, quantity)
	if err != nil {
		return "", err
	}

	payload := map[string]string{
		"currency":      formatChain(chain),
		"address":       address,
		"amount":        strconv.FormatFloat(quantity, 'f', 8, 64),
		"feeDeductType": "INTERNAL",
	}

	var obj *CreateWithdrawal
	err = c.do("POST", "/api/v1/withdrawals", payload, &obj, true)
	if err != nil {
		return "", err
	} else if obj.WithdrawalID == "" {
		return "", fmt.Errorf("empty withdrawal id")
	}

	return obj.WithdrawalID, nil
}

func (c *Client) GetWithdrawalByID(chain, withdrawalID string) (*types.Withdrawal, error) {
	payload := map[string]string{
		"currency": formatChain(chain),
	}

	var paginated *PaginatedResponse
	err := c.do("GET", "/api/v1/withdrawals", payload, &paginated, true)
	if err != nil {
		return nil, err
	}

	objs := make([]*Withdrawal, 0)
	err = json.Unmarshal(paginated.Items, &objs)
	if err != nil {
		return nil, err
	}

	for _, withdrawal := range objs {
		if withdrawal.ID == withdrawalID {
			var completed bool
			switch withdrawal.Status {
			case "SUCCESS":
				completed = true
			case "PROCESSING", "WALLET_PROCESSING":
			case "FAILURE":
				return nil, fmt.Errorf("withdrawal failed")
			default:
				return nil, fmt.Errorf("unknown withdrawal status %s", withdrawal.Status)
			}

			parsedWithdrawal := &types.Withdrawal{
				ID:        withdrawal.ID,
				TxID:      withdrawal.WalletTxId,
				Value:     withdrawal.Amount,
				Fee:       withdrawal.Fee,
				Completed: completed,
			}

			return parsedWithdrawal, nil
		}
	}

	return nil, fmt.Errorf("withdrawal not found")
}
