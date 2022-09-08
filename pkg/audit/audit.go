package audit

import (
	"time"

	"github.com/pkg/errors"

	"github.com/magicpool-co/pool/pkg/config"
	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
)

type Client struct {
	db    *db.DBClient
	nodes map[string]types.PayoutNode
}

func New(conf *config.Config) (*Client, error) {
	if conf.DB == nil {
		return nil, errors.New("New: DB is nil")
	} else if conf.PayoutNodes == nil {
		return nil, errors.New("New: PayoutNodes is nil")
	}

	client := &Client{
		db:    conf.DB,
		nodes: conf.PayoutNodes,
	}

	return client, nil
}

func (c *Client) CheckWallet(chain string) error {
	node, ok := c.nodes[chain]
	if !ok {
		return errors.Errorf("CheckWallet: unable to find node for %s", chain)
	}

	address, err := node.GetAddress(node.GetWallet().Address)
	if err != nil {
		return err
	}

	walletBalance := address.Balance
	utxoBalance, err := c.db.GetUnspentUTXOBalanceByCoin(chain)
	if err != nil {
		return err
	}

	payoutBalance, err := c.db.GetUnspentPayoutBalanceByCoin(chain)
	if err != nil {
		return err
	}

	feeBalance, err := c.db.GetUnspentPayoutFeeBalanceByCoin(chain)
	if err != nil {
		return err
	}

	pendingBalance, err := c.db.GetPendingBalanceByCoin(chain)
	if err != nil {
		return err
	}

	payoutBalance.Add(payoutBalance, feeBalance)
	payoutBalance.Add(payoutBalance, pendingBalance)
	if utxoBalance.Cmp(payoutBalance) != 0 {
		return errors.Errorf("CheckWallet: utxo and payout balance mismatch (%s): %s, %s", chain, utxoBalance, payoutBalance)
	}

	immatureBalance, err := c.db.GetImmatureRoundsBalanceByCoin(chain)
	if err != nil {
		return err
	}

	payoutBalance.Add(payoutBalance, immatureBalance)

	activeWithdrawals, err := c.db.GetActiveWithdrawalsByChain(chain)
	if err != nil {
		return err
	}

	unconfirmedPayouts, err := c.db.GetUnconfirmedPayoutsByChain(chain)
	if err != nil {
		return err
	}

	if len(activeWithdrawals) == 0 && len(unconfirmedPayouts) == 0 {
		if walletBalance.Cmp(payoutBalance) != 0 {
			return errors.Errorf("CheckWallet: wallet and payout balance mismatch (%s): %s, %s", chain, walletBalance, payoutBalance)
		}
	}

	return nil
}

// @TODO: audits the pending client's balance
// and makes sure the wallets are evenly matched
// with no deficit or excess balance
func (c *Client) CheckPending() error {
	return nil
}

// @TODO: audits a given round to verify that all
// balances were correctly handled and provides
// a trace of the fund movements
func (c *Client) CheckRound(id uint64) error {
	return nil
}

// @TODO: audits a given switch to verify that all
// balances were correctly handled and provides
// a trace of the fund movements
func (c *Client) CheckSwitch(id uint64) error {
	return nil
}

// @TODO audits a given user to verify that all
// historical and pending balances were correctly
// handled and provides a trace of the fund movements
func (c *Client) CheckMiner(interval time.Time) {

}

// @TODO audits a given recipient to verify that all
// historical and pending balances were correctly
// handled and provides a trace of the fund movements
func (c *Client) CheckRecipient(interval time.Time) {

}

// @TODO: generates a report for a given interval
// describing all revenue, fees, and profits along
// with a complete ledger of all fund movement
func (c *Client) GenerateReport(interval time.Time) {

}
