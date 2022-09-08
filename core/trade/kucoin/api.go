package kucoin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

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

func (c *Client) GetRate(base, quote string) (float64, error) {
	payload := map[string]string{
		"symbol": base + "-" + quote,
	}

	var obj *Symbol
	err := c.do("GET", "/api/v1/market/orderbook/level1", payload, &obj, false)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(obj.Price, 10)
}

func (c *Client) GetHistoricalRate(base, quote string, timestamp time.Time) (float64, error) {
	payload := map[string]string{
		"symbol":  base + "-" + quote,
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

/* wallet */

func (c *Client) GetWalletStatus(chain string) (bool, error) {
	var obj *Currency
	err := c.do("GET", "/api/v2/currencies/"+chain, nil, &obj, false)
	if err != nil {
		return false, err
	} else if !obj.IsDebitEnabled {
		return false, fmt.Errorf("trading is disabled for %s", chain)
	}

	for _, chainObj := range obj.Chains {
		if chainObj.ChainName == chain {
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

func (c *Client) getDepositAddresses(chain string) ([]*Address, error) {
	payload := map[string]string{
		"currency": chain,
	}

	objs := make([]*Address, 0)
	err := c.do("GET", "/api/v2/deposit-addresses", payload, &objs, true)

	return objs, err
}

func (c *Client) GetWalletAddress(chain string) (string, error) {
	var obj *Address
	addresses, err := c.getDepositAddresses(chain)
	if err != nil {
		return "", err
	} else if len(addresses) > 0 {
		obj = addresses[0]
	} else {
		// @TODO: do i need to use chain parameter for any currencies
		payload := map[string]string{
			"currency": chain,
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

func (c *Client) GetWalletBalance(chain string) (float64, error) {
	return 0, nil
}

/* deposit */

func (c *Client) GetDepositStatus(chain, txid string) (bool, error) {
	payload := map[string]string{
		"currency": chain,
	}

	var paginated *PaginatedResponse
	err := c.do("GET", "/api/v1/deposits", payload, &paginated, true)
	if err != nil {
		return false, err
	}

	objs := make([]*Deposit, 0)
	err = json.Unmarshal(paginated.Items, &objs)
	if err != nil {
		return false, err
	}

	for _, deposit := range objs {
		// @TODO: depending on the chain, txid may not actually match
		// like with ETH when they forward it via smart contract
		if deposit.WalletTxID == txid {
			switch deposit.Status {
			case "SUCCESS":
				return true, nil
			case "PROCESSING":
				return false, nil
			case "FAILURE":
				return false, fmt.Errorf("deposit failed")
			}
		}
	}

	return false, fmt.Errorf("deposit not found")
}

/* order */

func (c *Client) CreateOrder(base, quote, direction string, quantity float64) (string, error) {
	payload := map[string]string{
		"side":   strings.ToLower(direction),
		"symbol": base + "-" + quote,
		"type":   "market",
	}

	switch direction {
	case "BUY":
		payload["funds"] = strconv.FormatFloat(quantity, 'f', 8, 64)
	case "SELL":
		payload["size"] = strconv.FormatFloat(quantity, 'f', 8, 64)
	default:
		return "", fmt.Errorf("invalid trade direction %s", direction)
	}

	var obj *CreateOrder
	err := c.do("POST", "/api/v1/orders", payload, &obj, true)
	if err != nil {
		return "", err
	} else if obj.OrderID == "" {
		return "", fmt.Errorf("empty order id")
	}

	return obj.OrderID, nil
}

func (c *Client) GetOrderStatus(base, quote, orderID string) (bool, error) {
	var obj *Order
	err := c.do("GET", "/api/v1/orders/"+orderID, nil, &obj, true)
	if err != nil {
		return false, err
	} else if obj.CancelExist {
		return false, fmt.Errorf("order was cancelled")
	}

	return obj.IsActive, nil
}

/* withdrawal */

func (c *Client) checkWithdrawalQuota(chain string, quantity float64) error {
	payload := map[string]string{
		"currency": chain,
	}

	var obj *WithdrawalQuota
	err := c.do("GET", "/api/v1/withdrawals/quotas", payload, &obj, true)
	if err != nil {
		return err
	} else if !obj.IsWithdrawEnabled {
		return fmt.Errorf("withdrawals are not enabled for %s", chain)
	} else if remainingQuota, err := strconv.ParseFloat(obj.AvailableAmount, 10); err != nil {
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
		"currency":      chain,
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

func (c *Client) GetWithdrawalStatus(chain, withdrawalID string) (bool, error) {
	payload := map[string]string{
		"currency": chain,
	}

	var paginated *PaginatedResponse
	err := c.do("GET", "/api/v1/withdrawals", payload, &paginated, true)
	if err != nil {
		return false, err
	}

	objs := make([]*Withdrawal, 0)
	err = json.Unmarshal(paginated.Items, &objs)
	if err != nil {
		return false, err
	}

	for _, withdrawal := range objs {
		if withdrawal.ID == withdrawalID {
			switch withdrawal.Status {
			case "SUCCESS":
				return true, nil
			case "PROCESSING", "WALLET_PROCESSING":
				return false, nil
			case "FAILURE":
				return false, fmt.Errorf("withdrawal failed")
			}
		}
	}

	return false, fmt.Errorf("withdrawal not found")
}
