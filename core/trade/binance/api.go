package binance

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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
		"symbol": base + quote,
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
		"symbol":    base + quote,
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

/* wallet */

func (c *Client) GetWalletStatus(chain string) (bool, error) {
	payload := map[string]string{
		"asset": chain,
	}

	var obj *Asset
	err := c.do("GET", "/sapi/v1/asset/assetDetail", payload, &obj, securityTypeSigned)
	if err != nil {
		return false, err
	} else if !obj.DepositStatus {
		if obj.DepositTip != "" {
			return false, fmt.Errorf("deposits disabled for %s because: %s", chain, obj.DepositTip)
		}
		return false, fmt.Errorf("deposits disabled for %s", chain)
	} else if !obj.WithdrawStatus {
		return false, fmt.Errorf("withdrawals disabled for %s", chain)
	}

	return true, nil
}

func (c *Client) GetWalletAddress(chain string) (string, error) {
	// @TODO: does "network" need to be used for any coins we care about
	payload := map[string]string{
		"coin": chain,
	}

	var obj *Address
	err := c.do("GET", "/sapi/v1/capital/deposit/address", payload, &obj, securityTypeSigned)
	if err != nil {
		return "", err
	} else if obj.Address == "" {
		return "", fmt.Errorf("unable to find binance address for %s", chain)
	}

	return obj.Address, nil
}

func (c *Client) GetWalletBalance(chain string) (float64, error) {
	var obj *Account
	err := c.do("GET", "/api/v3/account", nil, &obj, securityTypeSigned)
	if err != nil {
		return 0, err
	}

	for _, balance := range obj.Balances {
		if chain == balance.Asset {
			return strconv.ParseFloat(balance.Free, 64)
		}
	}

	return 0, nil
}

/* deposit */

func (c *Client) GetDepositStatus(chain, txid string) (bool, error) {
	payload := map[string]string{
		"coin":  chain,
		"limit": "25",
	}

	objs := make([]*Deposit, 0)
	err := c.do("GET", "/sapi/v1/capital/deposit/hisre", payload, &objs, securityTypeSigned)
	if err != nil {
		return false, err
	}

	for _, deposit := range objs {
		// @TODO: depending on the chain, txid may not actually match
		// like with ETH when they forward it via smart contract
		if deposit.TxID == txid {
			switch deposit.Status {
			case 0: // pending
				return false, nil
			case 1: // success
				return true, nil
			case 6: // credited but cannot withdrawal
				return true, nil
			default:
				return false, fmt.Errorf("deposit %s has an unknown status status %d", txid, deposit.Status)
			}
		}
	}

	return false, fmt.Errorf("deposit not found")
}

/* order */

func (c *Client) CreateOrder(base, quote, direction string, quantity float64) (string, error) {
	payload := map[string]string{
		"symbol":           base + quote,
		"side":             direction,
		"type":             "MARKET",
		"newOrderRespType": "RESULT",
	}

	switch direction {
	case "BUY":
		payload["quoteOrderQty"] = strconv.FormatFloat(quantity, 'f', 8, 64)
	case "SELL":
		payload["quantity"] = strconv.FormatFloat(quantity, 'f', 8, 64)
	default:
		return "", fmt.Errorf("invalid trade direction %s", direction)
	}

	var obj *Order
	err := c.do("POST", "/api/v3/order", payload, &obj, securityTypeSigned)
	if err != nil {
		return "", err
	} else if obj.ClientOrderID == "" {
		return "", fmt.Errorf("order for %s-%s has no ClientOrderID", base, quote)
	}

	return obj.ClientOrderID, nil
}

func (c *Client) GetOrderStatus(base, quote, orderID string) (bool, error) {
	payload := map[string]string{
		"symbol":            base + quote,
		"origClientOrderId": orderID,
	}

	var obj *Order
	err := c.do("GET", "/api/v3/order", payload, &obj, securityTypeSigned)
	if err != nil {
		return false, err
	}

	switch obj.Status {
	case "NEW", "PARTIALLY_FILLED":
		return false, nil
	case "FILLED":
		return true, nil
	case "PENDING_CANCEL", "CANCELLED":
		return false, fmt.Errorf("order %s was cancelled", orderID)
	case "REJECTED", "EXPIRED":
		return false, fmt.Errorf("order %s %s", orderID, strings.ToLower(obj.Status))
	default:
		return false, fmt.Errorf("order %s has an unknown status %s", orderID, obj.Status)
	}
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
		return "", fmt.Errorf("withdrawal for %s has no ID", chain)
	}

	return obj.ID, nil
}

func (c *Client) GetWithdrawalStatus(chain, withdrawalID string) (bool, error) {
	payload := map[string]string{
		"coin":            chain,
		"withdrawOrderId": withdrawalID,
		"limit":           "25",
	}

	objs := make([]*Withdrawal, 0)
	err := c.do("GET", "/sapi/v1/capital/withdraw/history", payload, &objs, securityTypeSigned)
	if err != nil {
		return false, err
	} else if len(objs) == 0 {
		return false, fmt.Errorf("withdrawal not found")
	} else if len(objs) > 1 {
		return false, fmt.Errorf("more than 1 withdrawal found with id %s", withdrawalID)
	}

	status := objs[0].Status
	switch status {
	case 0: // email sent
		return false, fmt.Errorf("withdrawal %s is waiting for an email", withdrawalID)
	case 1: // cancelled
		return false, fmt.Errorf("withdrawal %s was cancelled", withdrawalID)
	case 2: // awaiting approval
		return false, fmt.Errorf("withdrawal %s is awaiting approval", withdrawalID)
	case 3: // rejected
		return false, fmt.Errorf("withdrawal %s was rejected", withdrawalID)
	case 4: // processing
		return false, nil
	case 5: // failure
		return false, fmt.Errorf("withdrawal %s has failed", withdrawalID)
	case 6: // completed
		return true, nil
	default:
		return false, fmt.Errorf("withdrawal %s has an unknown status status %d", withdrawalID, status)
	}
}
