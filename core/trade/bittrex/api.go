package bittrex

import (
	"fmt"
	"strconv"
	"time"

	"github.com/goccy/go-json"
)

/* account */

func (c *Client) GetAccountStatus() error {
	return nil
}

/* rate */

func (c *Client) GetRate(base, quote string) (float64, error) {
	res, err := c.do("GET", "/markets/"+base+"-"+quote+"/ticker", "", false)
	if err != nil {
		return 0.0, err
	}

	obj := new(RateResponse)
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return 0.0, err
	}

	return strconv.ParseFloat(obj.LastTradeRate, 64)
}

func (c *Client) GetHistoricalRate(base, quote string, timestamp time.Time) (float64, error) {
	path := "markets/" + base + "-" + quote + "/candles/MINUTE_5"
	if time.Since(timestamp) < time.Hour*24 {
		path += "/recent"
	} else {
		path += fmt.Sprintf("/historical/%d/%d/%d", timestamp.Year(), timestamp.Month(), timestamp.Day())
	}

	res, err := c.do("GET", path, "", false)
	if err != nil {
		return 0.0, err
	}

	obj := make([]*HistoricalRateResponse, 0)
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return 0.0, err
	}

	// @TODO: this logic is messy at best
	closestDiff := time.Hour * 24 * 365
	var closestRate float64
	for _, item := range obj {
		var replace bool

		diff1 := item.StartsAt.Sub(timestamp)
		diff2 := timestamp.Sub(item.StartsAt)
		if diff1 < 0 {
			diff1 = -1 * diff1
		}

		if diff1 < closestDiff {
			closestDiff = diff1
			replace = true
		}

		if diff2 < 0 {
			diff2 = -1 * diff2
		}

		if diff2 < closestDiff {
			closestDiff = diff2
			replace = true
		}

		if replace {
			closestRate, err = strconv.ParseFloat(item.Close, 64)
			if err != nil {
				return 0.0, err
			}
		}
	}

	return closestRate, nil
}

/* wallet */

func (c *Client) GetWalletStatus(chain string) (bool, error) {
	res, err := c.do("GET", "/currencies/"+chain, "", false)
	if err != nil {
		return false, err
	}

	currency := new(CurrencyV3)
	err = json.Unmarshal(res, currency)
	if err != nil {
		return false, err
	}

	return currency.Status == "ONLINE", nil
}

func (c *Client) GetWalletAddress(chain string) (string, error) {
	res, err := c.do("GET", "/addresses/"+chain, "", true)
	if err != nil {
		return "", err
	}

	address := new(AddressV3)
	err = json.Unmarshal(res, address)
	if err != nil {
		return "", err
	}

	// create the address if it doesn't exist yet
	if address.CryptoAddress == "" {
		payload, err := json.Marshal(AddressParams{CurrencySymbol: chain})
		if err != nil {
			return "", err
		}

		_, err = c.do("POST", "/addresses", string(payload), true)
		if err != nil {
			return "", err
		}

		res, err = c.do("GET", "/addresses/"+chain, "", true)
		if err != nil {
			return "", err
		}

		err = json.Unmarshal(res, address)
		if err != nil {
			return "", err
		}
	}

	if address.CryptoAddress == "" {
		return "", fmt.Errorf("unable to generate bittrex address for %s", chain)
	}

	return address.CryptoAddress, nil
}

func (c *Client) GetWalletBalance(chain string) (float64, error) {
	res, err := c.do("GET", "/balances/"+chain, "", true)
	if err != nil {
		return 0.0, err
	}

	balance := new(Balance)
	err = json.Unmarshal(res, &balance)
	if err != nil {
		return 0.0, err
	}

	return strconv.ParseFloat(balance.Available, 64)
}

/* deposit */

func (c *Client) GetDepositStatus(chain, txid string) (bool, error) {
	res, err := c.do("GET", "deposits/ByTxId/"+txid, "", true)
	if err != nil {
		return false, err
	}

	deposits := make([]*DepositV3, 0)
	err = json.Unmarshal(res, &deposits)
	if err != nil {
		return false, err
	} else if len(deposits) == 0 {
		return false, fmt.Errorf("deposit not found")
	} else if len(deposits) > 1 {
		return false, fmt.Errorf("more than 1 deposit found for txid %s", txid)
	}

	switch deposits[0].Status {
	case "PENDING":
		return false, nil
	case "COMPLETED":
		return true, nil
	case "ORPHANED", "INVALIDATED":
		return false, fmt.Errorf("deposit %s was %s", txid, deposits[0].Status)
	default:
		return false, fmt.Errorf("deposit %s has an unknown status %s", txid, deposits[0].Status)
	}
}

/* order */

func (c *Client) CreateOrder(base, quote string, direction string, quantity float64) (string, error) {
	var params CreateOrderParams
	switch direction {
	case "BUY":
		params = CreateOrderParams{
			Type:         "CEILING_MARKET",
			MarketSymbol: base + "-" + quote,
			Direction:    "BUY",
			TimeInForce:  "FILL_OR_KILL",
			Ceiling:      quantity,
		}
	case "SELL":
		params = CreateOrderParams{
			Type:         "MARKET",
			MarketSymbol: base + "-" + quote,
			Direction:    "BUY",
			TimeInForce:  "FILL_OR_KILL",
			Quantity:     quantity,
		}
	default:
		return "", fmt.Errorf("invalid trade direction %s", direction)
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	res, err := c.do("POST", "/orders/", string(payload), true)
	if err != nil {
		return "", err
	}

	order := new(OrderV3)
	err = json.Unmarshal(res, &order)

	return order.ID, err
}

func (c *Client) GetOrderStatus(base, quote, orderID string) (bool, error) {
	res, err := c.do("GET", "/orders/"+orderID, "", true)
	if err != nil {
		return false, err
	}

	obj := new(OrderV3)
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return false, err
	}

	switch obj.Status {
	case "OPEN":
		return false, nil
	case "CLOSED":
		return true, nil
	default:
		return false, fmt.Errorf("order %s has an unknown status %s", orderID, obj.Status)
	}
}

/* withdrawal */

func (c *Client) CreateWithdrawal(chain, address string, quantity float64) (string, error) {
	params := WithdrawalParams{
		CurrencySymbol: chain,
		Quantity:       quantity,
		CryptoAddress:  address,
	}

	payload, err := json.Marshal(params)
	res, err := c.do("POST", "/withdrawals", string(payload), true)
	if err != nil {
		return "", err
	}

	obj := new(WithdrawalV3)
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return "", err
	}

	return obj.ID, nil
}

func (c *Client) GetWithdrawalStatus(chain, withdrawalID string) (bool, error) {
	res, err := c.do("GET", "/withdrawals/"+withdrawalID, "", true)
	if err != nil {
		return false, err
	}

	obj := new(WithdrawalV3)
	err = json.Unmarshal(res, obj)
	if err != nil {
		return false, err
	}

	switch obj.Status {
	case "REQUESTED", "AUTHORIZED", "PENDING":
		return false, nil
	case "COMPLETED":
		return true, nil
	case "CANCELLED":
		return false, fmt.Errorf("withdrawal %s was cancelled", withdrawalID)
	case "ERROR_INVALID_ADDRESS":
		return false, fmt.Errorf("withdrawal %s failed due to an invalid address", withdrawalID)
	default:
		return false, fmt.Errorf("withdrawal %s has an unknown status status %s", withdrawalID, obj.Status)
	}
}
