package accounting

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/bittrex"
	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
	"github.com/magicpool-co/pool/pkg/utils"
)

var switchThresholds = map[string]*big.Int{
	"ETH":  utils.MustParseBigInt("1000000000000000000"),
	"ETC":  utils.MustParseBigInt("150000000000000000000"),
	"RVN":  utils.MustParseBigInt("2000000000000"),
	"BTC":  utils.MustParseBigInt("10000000"),
	"USDC": utils.MustParseBigInt("5000000000"),
}

type SwitchAccountant struct {
	bimap       *bimap
	rates       map[string]map[string]float64
	units       map[string]*big.Int
	payouts     map[uint64]*db.Payout
	payoutIDs   map[uint64]bool
	withdrawals map[string]*db.SwitchWithdrawal
}

func NewSwitchAccountant() *SwitchAccountant {
	return &SwitchAccountant{
		bimap:       NewBimap(),
		payouts:     make(map[uint64]*db.Payout),
		payoutIDs:   make(map[uint64]bool),
		withdrawals: make(map[string]*db.SwitchWithdrawal),
	}
}

/* add methods */

func (a *SwitchAccountant) AddRates(rates map[string]map[string]float64) {
	a.rates = rates
}

func (a *SwitchAccountant) AddUnits(units map[string]*big.Int) {
	a.units = units
}

/* read methods */

func (a *SwitchAccountant) InputValue(coin string) *big.Int {
	sum, ok := a.bimap.GetInputSum(coin)
	if !ok {
		return new(big.Int)
	}

	return sum
}

func (a *SwitchAccountant) InputValues() map[string]*big.Int {
	values := make(map[string]*big.Int)
	for _, coin := range a.bimap.GetInputKeys() {
		values[coin], _ = a.bimap.GetInputSum(coin)
	}

	return values
}

func (a *SwitchAccountant) Balances() []*db.Balance {
	return a.bimap.GetBalances()
}

func (a *SwitchAccountant) Payouts() []*db.Payout {
	payouts := make([]*db.Payout, 0)
	for _, payout := range a.payouts {
		payouts = append(payouts, payout)
	}

	return payouts
}

func (a *SwitchAccountant) PayoutIDs() []uint64 {
	payoutIDs := make([]uint64, 0)
	for id := range a.payoutIDs {
		payoutIDs = append(payoutIDs, id)
	}

	return payoutIDs
}

func (a *SwitchAccountant) Withdrawals() []*db.SwitchWithdrawal {
	withdrawals := make([]*db.SwitchWithdrawal, 0)
	for _, withdrawal := range a.withdrawals {
		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals
}

/* assistant abstraction */

func (a *SwitchAccountant) AddBalance(balance *db.Balance, estimate bool) error {
	if !balance.InValue.Valid {
		return fmt.Errorf("invalid InValue for value %d", balance.ID)
	}

	value := balance.InValue.BigInt
	estimatedValue := new(big.Int)
	base := balance.InCoin
	quote := balance.OutCoin

	if estimate {
		rawRate, ok := a.rates[base][quote]
		if !ok {
			return fmt.Errorf("accountant has no rate for %s-%s", base, quote)
		}

		baseUnits, ok := a.units[base]
		if !ok {
			return fmt.Errorf("accountant has no units for base %s", base)
		}

		quoteUnits, ok := a.units[quote]
		if !ok {
			return fmt.Errorf("accountant has no units for quote %s", quote)
		}

		rawRateStr := fmt.Sprintf("%.8f", rawRate)
		rate, err := utils.StringDecimalToBigint(rawRateStr, quoteUnits)
		if err != nil {
			return err
		}

		estimatedValue.Mul(value, rate)
		estimatedValue.Div(estimatedValue, baseUnits)
	}

	a.bimap.AddByInput(balance, estimatedValue)

	if _, ok := a.payoutIDs[balance.OutPayoutID]; !ok {
		a.payoutIDs[balance.OutPayoutID] = true
	}

	return nil
}

func (a *SwitchAccountant) CheckThresholds() error {
	if a.rates == nil {
		return fmt.Errorf("rates are not set")
	} else if a.units == nil {
		return fmt.Errorf("units are not set")
	}

	iterations := len(a.bimap.GetInputKeys())
	if outputLength := len(a.bimap.GetOutputKeys()); outputLength > iterations {
		iterations = outputLength
	}

	for i := 0; i < iterations; i++ {
		changed := false
		for _, inputCoin := range a.bimap.GetInputKeys() {
			threshold, ok := switchThresholds[inputCoin]
			if !ok {
				return fmt.Errorf("unable to find threshold for %s", inputCoin)
			}

			v, ok := a.bimap.GetInputSum(inputCoin)
			if !ok {
				return fmt.Errorf("unable to find input sum for %s", inputCoin)
			}

			if v.Cmp(threshold) < 0 {
				changed = true
				a.bimap.DeleteByInput(inputCoin)
			}
		}

		for _, outputCoin := range a.bimap.GetOutputKeys() {
			threshold, ok := switchThresholds[outputCoin]
			if !ok {
				return fmt.Errorf("unable to find threshold for %s", outputCoin)
			}

			v, ok := a.bimap.GetOutputSum(outputCoin)
			if !ok {
				return fmt.Errorf("unable to find output sum for %s", outputCoin)
			}

			if v.Cmp(threshold) < 0 {
				changed = true
				a.bimap.DeleteByOutput(outputCoin)
			}
		}

		if !changed {
			break
		}
	}

	return nil
}

func (a *SwitchAccountant) AddDeposit(deposit *db.SwitchDeposit) error {
	if err := a.bimap.AddDeposit(deposit); err != nil {
		return err
	}

	depositValue := new(big.Int).Add(deposit.Value.BigInt, deposit.Fees.BigInt)
	inputValue, ok := a.bimap.GetInputSum(deposit.CoinID)
	if !ok {
		return fmt.Errorf("unable to find input sum for %s", deposit.CoinID)
	}

	if inputValue.Cmp(depositValue) != 0 {
		return fmt.Errorf("input mismatch on deposit %d", deposit.ID)
	}

	return nil
}

func (a *SwitchAccountant) GenerateTrades(switchID uint64) ([][]*db.SwitchTrade, error) {
	var pathID uint64 = 0
	allTrades := make([][]*db.SwitchTrade, 0)

	for _, base := range a.bimap.GetInputKeys() {
		for _, quote := range a.bimap.GetOutputKeys() {
			entry, ok := a.bimap.GetByInput(base, quote)
			if !ok {
				continue
			}

			path, ok := bittrex.TradePaths[base][quote]
			if !ok {
				return nil, fmt.Errorf("unsupported path %s-%s", base, quote)
			}

			trades := make([]*db.SwitchTrade, len(path))
			// iterate backwards through the path
			for i := len(path) - 1; i >= 0; i-- {
				singleTradeValue := big.NewInt(0)
				if i == 0 {
					singleTradeValue = entry.tradeValue
				}

				var fromCoin, toCoin string
				var direction int
				if path[i].Direction == bittrex.SELL {
					fromCoin = path[i].Base
					toCoin = path[i].Quote
					direction = int(types.SellDirection)
				} else {
					fromCoin = path[i].Quote
					toCoin = path[i].Base
					direction = int(types.BuyDirection)
				}

				trades[i] = &db.SwitchTrade{
					PathID: pathID,
					Stage:  i + 1,

					SwitchID:  switchID,
					FromCoin:  fromCoin,
					ToCoin:    toCoin,
					Market:    path[i].Market,
					Direction: direction,

					Value: db.NullBigInt{singleTradeValue, true},

					Initiated: false,
					Open:      false,
					Filled:    false,
				}
			}

			allTrades = append(allTrades, trades)
			pathID += 1
		}
	}

	return allTrades, nil
}

func (a *SwitchAccountant) AddPayout(payout *db.Payout) error {
	if _, ok := a.payouts[payout.ID]; ok {
		return fmt.Errorf("payout %d is already added", payout.ID)
	}

	a.payouts[payout.ID] = payout

	return nil
}

func (a *SwitchAccountant) AddInitialPath(switchID uint64, path []*db.SwitchTrade) error {
	for i, trade := range path {
		if trade.Stage != i+1 {
			return fmt.Errorf("out of order trade %d", trade.ID)
		} else if !trade.Fees.Valid {
			return fmt.Errorf("invalid fees for trade %d", trade.ID)
		} else if !trade.Value.Valid {
			return fmt.Errorf("invalid value for trade %d", trade.ID)
		} else if !trade.Proceeds.Valid {
			return fmt.Errorf("invalid proceeds for trade %d", trade.ID)
		}
	}

	cumulativeFees := new(big.Int)
	for i := 0; i < len(path); i++ {
		if i > 0 {
			inValue := path[i].Value.BigInt
			outValue := new(big.Int).Add(path[i].Proceeds.BigInt, path[i].Fees.BigInt)
			cumulativeFees.Mul(cumulativeFees, outValue)
			cumulativeFees.Div(cumulativeFees, inValue)
		}

		cumulativeFees.Add(cumulativeFees, path[i].Fees.BigInt)
	}

	finalTrade := path[len(path)-1]
	finalTradeValue := finalTrade.Proceeds.BigInt
	if withdrawal, ok := a.withdrawals[finalTrade.ToCoin]; !ok {
		a.withdrawals[finalTrade.ToCoin] = &db.SwitchWithdrawal{
			SwitchID: switchID,
			CoinID:   finalTrade.ToCoin,

			Value:     db.NullBigInt{finalTradeValue, true},
			TradeFees: db.NullBigInt{cumulativeFees, true},

			Confirmed: false,
			Spent:     false,
		}
	} else {
		withdrawal.Value.BigInt.Add(withdrawal.Value.BigInt, finalTradeValue)
		withdrawal.TradeFees.BigInt.Add(withdrawal.TradeFees.BigInt, cumulativeFees)
	}

	return nil
}

func (a *SwitchAccountant) AddFinalPath(path []*db.SwitchTrade) error {
	if err := a.bimap.AddTradePath(path[0], path[len(path)-1]); err != nil {
		return err
	}

	return nil
}

func (a *SwitchAccountant) AddWithdrawal(withdrawal *db.SwitchWithdrawal) error {
	if err := a.bimap.AddWithdrawal(withdrawal); err != nil {
		return err
	}

	return nil
}

func (a *SwitchAccountant) Distribute() error {
	if err := a.bimap.DistributeToBalances(); err != nil {
		return err
	}

	for _, balance := range a.bimap.GetBalances() {
		finalValue := balance.OutValue.BigInt
		finalExchangeFees := balance.ExchangeFees.BigInt
		finalPoolFees := balance.PoolFees.BigInt

		payout, ok := a.payouts[balance.OutPayoutID]
		if !ok {
			return fmt.Errorf("unable to find payout %d", balance.OutPayoutID)
		}

		if payout.CoinID == balance.OutCoin {
			payout.Value.BigInt.Add(payout.Value.BigInt, finalValue)
		} else if payout.FeeBalanceCoin == balance.OutCoin {
			if !payout.FeeBalancePending {
				return fmt.Errorf("payout %d has extra fee balance", payout.ID)
			} else {
				payout.FeeBalancePending = false
				payout.InFeeBalance.BigInt.Add(payout.InFeeBalance.BigInt, finalValue)
			}
		} else {
			return fmt.Errorf("mismatch: payout %d and balance %d", payout.ID, balance.ID)
		}

		payout.ExchangeFees.BigInt.Add(payout.ExchangeFees.BigInt, finalExchangeFees)
		// @TODO: make this in USDC if its for fee balance
		payout.PoolFees.BigInt.Add(payout.PoolFees.BigInt, finalPoolFees)
	}

	return nil
}
