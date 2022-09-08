package trade

import (
	"fmt"
	"time"

	"github.com/magicpool-co/pool/core/trade/binance"
	"github.com/magicpool-co/pool/core/trade/bittrex"
	"github.com/magicpool-co/pool/core/trade/kucoin"
)

type Exchange interface {
	GetAccountStatus() error
	GetRate(base, quote string) (float64, error)
	GetHistoricalRate(base, quote string, timestamp time.Time) (float64, error)
	GetWalletStatus(chain string) (bool, error)
	GetWalletAddress(chain string) (string, error)
	GetWalletBalance(chain string) (float64, error)
	GetDepositStatus(chain, txid string) (bool, error)
	CreateOrder(base, quote, direction string, quantity float64) (string, error)
	GetOrderStatus(base, quote, orderID string) (bool, error)
	CreateWithdrawal(chain, address string, quantity float64) (string, error)
	GetWithdrawalStatus(chain, withdrawalID string) (bool, error)
}

type Client struct {
	exchange Exchange
}

func New(exchangeName, apiKey, secretKey, secretPassphrase string) (*Client, error) {
	var exchange Exchange
	switch exchangeName {
	case "binance":
		exchange = binance.New(apiKey, secretKey)
	case "bittrex":
		exchange = bittrex.New(apiKey, secretKey)
	case "kucoin":
		exchange = kucoin.New(apiKey, secretKey, secretPassphrase)
	default:
		return nil, fmt.Errorf("unsupported exchange %s", exchangeName)
	}

	client := &Client{
		exchange: exchange,
	}

	return client, nil
}

func (c *Client) InitiateDeposits() error {
	for _, chain := range []string{} {
		walletActive, err := c.exchange.GetWalletStatus(chain)
		if err != nil {
			return err
		} else if !walletActive {
			return fmt.Errorf("deposits not enabled for chain %s", chain)
		}

		walletAddress, err := c.exchange.GetWalletAddress(chain)
		if err != nil {
			return err
		}

		// @TODO: send deposit to wallet address
		fmt.Println(walletAddress)
	}

	return nil
}

func (c *Client) FinalizeDeposits() error {
	for _, chain := range []string{} {
		// @TODO: get txid for chains
		var txid string
		depositCompleted, err := c.exchange.GetDepositStatus(chain, txid)
		if err != nil {
			return err
		} else if !depositCompleted {
			continue
		}

		// @TODO: finalize deposit in db
	}

	return nil
}

func (c *Client) InitiateTrades() error {
	return nil
}

func (c *Client) ExecuteTradeStage() error {
	return nil
}

func (c *Client) InitiateWithdrawals() error {
	for _, chain := range []string{} {
		// @TODO: properly manage nodes w/ address
		var address string
		withdrawalID, err := c.exchange.CreateWithdrawal(chain, address, 0.0)
		if err != nil {
			return err
		}

		// @TODO: insert withdrawal
		fmt.Println(withdrawalID)
	}

	return nil
}

func (c *Client) ConfirmWithdrawals() error {
	for _, chain := range []string{} {
		// @TODO: get withdrawalID for chains
		var withdrawalID string
		withdrawalCompleted, err := c.exchange.GetWithdrawalStatus(chain, withdrawalID)
		if err != nil {
			return err
		} else if !withdrawalCompleted {
			continue
		}
	}

	return nil
}

func (c *Client) CreditWithdrawals() error {
	return nil
}
