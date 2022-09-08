package trader

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/accounting"
	"github.com/magicpool-co/pool/pkg/bittrex"
	"github.com/magicpool-co/pool/pkg/config"
	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/telegram"
	"github.com/magicpool-co/pool/pkg/types"

	"github.com/pkg/errors"
)

type client struct {
	db       *db.DBClient
	bittrex  bittrex.BittrexClient
	nodes    map[string]types.PayoutNode
	telegram *telegram.Config
}

func initClient(conf *config.Config) (*client, error) {
	if conf.DB == nil {
		return nil, errors.New("no database object in the config")
	} else if err := conf.DB.Ping(); err != nil {
		return nil, errors.Wrap(err, "unable to connect to database")
	}

	c := &client{
		db:       conf.DB,
		bittrex:  conf.Bittrex,
		nodes:    conf.PayoutNodes,
		telegram: conf.Telegram,
	}

	return c, nil
}

func SwitchBalances(conf *config.Config) error {
	if c, err := initClient(conf); err != nil {
		return err
	} else {
		if err := c.initiateSwitch(); err != nil {
			return err
		}

		if switches, err := c.db.GetActiveSwitches(); err != nil {
			return err
		} else {
			for _, profitSwitch := range switches {
				if err := c.handleSwitch(profitSwitch); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *client) handleSwitch(profitSwitch *db.Switch) error {
	if profitSwitch.Status >= int(types.SwitchComplete) {
		return nil
	}

	var err error
	status := types.SwitchStatus(profitSwitch.Status)

	switch status {
	case types.DepositInactive:
		err = c.initiateDeposits(profitSwitch)
	case types.DepositUnregistered:
		err = c.registerDeposits(profitSwitch)
	case types.DepositUnconfirmed:
		err = c.confirmDeposits(profitSwitch)
	case types.DepositComplete:
		err = c.initiateTrades(profitSwitch)
	case types.TradesInactive:
		err = c.executeTradeStage(profitSwitch, 1)
	case types.TradesActiveStageOne:
		return fmt.Errorf("deprecated stage %d for %d", status, profitSwitch.ID)
	case types.TradesCompleteStageOne:
		err = c.executeTradeStage(profitSwitch, 2)
	case types.TradesActiveStageTwo:
		return fmt.Errorf("deprecated stage %d for %d", status, profitSwitch.ID)
	case types.TradesCompleteStageTwo:
		err = c.initiateWithdrawals(profitSwitch)
	case types.WithdrawalsActive:
		err = c.confirmWithdrawals(profitSwitch)
	case types.WithdrawalsComplete:
		err = c.creditWithdrawals(profitSwitch)
	default:
		return fmt.Errorf("invalid switch status %d for switch %d", status, profitSwitch.ID)
	}

	return err
}

func (c *client) initiateSwitch() error {
	if activeSwitches, err := c.db.GetActiveSwitches(); err != nil {
		return err
	} else if len(activeSwitches) > 0 {
		return nil
	}

	balances, err := c.db.GetUnswitchedBalances()
	if err != nil {
		return err
	}

	rates := make(map[string]map[string]float64)
	units := make(map[string]*big.Int)
	enabled := make(map[string]bool)
	for _, balance := range balances {
		base := balance.InCoin
		quote := balance.OutCoin

		if _, ok := enabled[base]; !ok {
			currency, err := c.bittrex.GetCurrency(base)
			if err != nil {
				return err
			}
			enabled[base] = currency.Status == "ONLINE"
		}

		if _, ok := rates[base]; !ok {
			rates[base] = make(map[string]float64)
		}

		if _, ok := rates[base][quote]; !ok {
			rate, err := c.bittrex.GetRate(base, quote)
			if err != nil {
				return err
			}

			rates[base][quote] = rate
		}

		for _, coin := range []string{base, quote} {
			if _, ok := units[coin]; !ok {
				node, ok := c.nodes[coin]
				if !ok {
					return fmt.Errorf("unable to find node for %s", coin)
				}

				units[coin] = node.GetUnits().Big()
			}
		}
	}

	accountant := accounting.NewSwitchAccountant()
	accountant.AddRates(rates)
	accountant.AddUnits(units)

	for _, balance := range balances {
		if !enabled[balance.InCoin] {
			continue
		} else if err := accountant.AddBalance(balance, true); err != nil {
			return err
		}
	}

	if err := accountant.CheckThresholds(); err != nil {
		return err
	} else if len(accountant.Balances()) == 0 {
		return nil
	}

	profitSwitch := &db.Switch{
		Status: int(types.DepositInactive),
	}

	if switchID, err := c.db.InsertSwitch(profitSwitch); err != nil {
		return err
	} else {
		for _, balance := range accountant.Balances() {
			balance.SwitchID = sql.NullInt64{int64(switchID), true}
			cols := []string{"switch_id"}
			if err := c.db.UpdateBalance(balance, cols); err != nil {
				return err
			}
		}

		if c.telegram != nil {
			c.telegram.NotifyInitiateSwitch(switchID)
		}
	}

	return nil
}
