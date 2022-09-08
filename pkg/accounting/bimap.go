package accounting

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/magicpool-co/pool/pkg/db"
)

type bimapEntry struct {
	balances       []*db.Balance
	inputValue     *big.Int
	tradeValue     *big.Int
	depositFees    *big.Int
	tradeFees      *big.Int
	withdrawalFees *big.Int
	actualOutput   *big.Int
	estimateOutput *big.Int
	finalOutput    *big.Int
}

func newBimapEntry(balance *db.Balance, inVal, outVal *big.Int) *bimapEntry {
	var estimatedOutput *big.Int
	if outVal != nil {
		estimatedOutput = new(big.Int).Set(outVal)
	}

	e := &bimapEntry{
		balances:       []*db.Balance{balance},
		inputValue:     new(big.Int).Set(inVal),
		estimateOutput: estimatedOutput,
	}

	return e
}

func (e *bimapEntry) Add(entry *bimapEntry) {
	e.inputValue.Add(e.inputValue, entry.inputValue)
	e.estimateOutput.Add(e.estimateOutput, entry.estimateOutput)
	e.balances = append(e.balances, entry.balances...)
}

type bimap struct {
	inputSums  map[string]*big.Int
	outputSums map[string]*big.Int
	inputs     map[string]map[string]*bimapEntry
	outputs    map[string]map[string]*bimapEntry
}

func NewBimap() *bimap {
	m := &bimap{
		inputs:  make(map[string]map[string]*bimapEntry),
		outputs: make(map[string]map[string]*bimapEntry),
	}

	return m
}

/* aggregation reads */

func (m *bimap) GetInputKeys() []string {
	keys := make([]string, 0)
	for key := range m.inputs {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (m *bimap) GetOutputKeys() []string {
	keys := make([]string, 0)
	for key := range m.outputs {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (m *bimap) GetBalances() []*db.Balance {
	balances := make([]*db.Balance, 0)
	for inputKey := range m.inputs {
		for _, entry := range m.inputs[inputKey] {
			balances = append(balances, entry.balances...)
		}
	}

	sort.Slice(balances, func(i, j int) bool {
		return balances[i].ID < balances[j].ID
	})

	return balances
}

/* general writes */

func (m *bimap) AddByInput(balance *db.Balance, estimateOutput *big.Int) error {
	if !balance.InValue.Valid {
		return fmt.Errorf("invalid InValue for value %d", balance.ID)
	}

	inputValue := balance.InValue.BigInt
	inputKey := balance.InCoin
	outputKey := balance.OutCoin

	if _, ok := m.inputs[inputKey]; !ok {
		m.inputs[inputKey] = make(map[string]*bimapEntry)
	}

	if _, ok := m.outputs[outputKey]; !ok {
		m.outputs[outputKey] = make(map[string]*bimapEntry)
	}

	entry := newBimapEntry(balance, inputValue, estimateOutput)

	_, inputOk := m.inputs[inputKey][outputKey]
	_, outputOk := m.outputs[outputKey][inputKey]

	if inputOk != outputOk {
		return fmt.Errorf("mismatch: input %s, output %s", inputKey, outputKey)
	} else if inputOk && outputOk {
		m.inputs[inputKey][outputKey].Add(entry)
	} else {
		m.inputs[inputKey][outputKey] = entry
		m.outputs[outputKey][inputKey] = entry
	}

	return nil
}

func (m *bimap) AddDeposit(deposit *db.SwitchDeposit) error {
	if !deposit.Value.Valid {
		return fmt.Errorf("invalid Value for deposit %d", deposit.ID)
	} else if !deposit.Fees.Valid {
		return fmt.Errorf("invalid depositFees for deposit %d", deposit.ID)
	}

	inputKey := deposit.CoinID
	depositValue := deposit.Value.BigInt
	depositFees := deposit.Fees.BigInt

	inputs, ok := m.inputs[inputKey]
	if !ok {
		return fmt.Errorf("unable to find input %s", inputKey)
	}

	totalInputValue := new(big.Int)
	for _, entry := range inputs {
		totalInputValue.Add(totalInputValue, entry.inputValue)
	}

	calculatedFees := new(big.Int).Sub(totalInputValue, depositValue)
	if calculatedFees.Cmp(depositFees) != 0 {
		return fmt.Errorf("fee mismatch on deposit %d", deposit.ID)
	}

	usedValue := new(big.Int)
	usedFees := new(big.Int)
	for _, entry := range inputs {
		entry.tradeValue = adjustValue(depositValue, entry.inputValue, totalInputValue)
		usedValue.Add(usedValue, entry.tradeValue)

		entry.depositFees = new(big.Int).Sub(entry.inputValue, entry.tradeValue)
		usedFees.Add(usedFees, entry.depositFees)
	}

	// distribute remainders
	valueRemainder := new(big.Int).Sub(depositValue, usedValue)
	feeRemainder := new(big.Int).Sub(depositFees, usedFees)

	// since inputs is a map order can change, sort the keys
	// to make sure remainder distribution is deterministic
	keys := make([]string, 0)
	for key := range inputs {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	if len(keys) == 0 {
		return fmt.Errorf("no entries to distribute the input remainder to")
	} else {
		entry := inputs[keys[0]]
		entry.tradeValue.Add(entry.tradeValue, valueRemainder)
		entry.depositFees.Sub(entry.depositFees, feeRemainder)
	}

	return nil
}

func (m *bimap) AddTradePath(initial, final *db.SwitchTrade) error {
	if !initial.Value.Valid {
		return fmt.Errorf("invalid value for final trade %d", initial.ID)
	} else if !final.Proceeds.Valid {
		return fmt.Errorf("invalid proceeds for final trade %d", final.ID)
	} else if !final.Fees.Valid {
		return fmt.Errorf("invalid fees for final trade %d", final.ID)
	}

	inCoin := initial.FromCoin
	outCoin := final.ToCoin

	if entry, ok := m.inputs[inCoin][outCoin]; !ok {
		return fmt.Errorf("unable to find input [%s][%s]", inCoin, outCoin)
	} else if initial.Value.BigInt.Cmp(entry.tradeValue) != 0 {
		return fmt.Errorf("mismatch trade value for trade %d (%s, %s)", initial.ID, initial.Value.BigInt, entry.tradeValue)
	} else {
		entry.actualOutput = final.Proceeds.BigInt
		entry.tradeFees = final.Fees.BigInt
	}

	return nil
}

func (m *bimap) AddWithdrawal(withdrawal *db.SwitchWithdrawal) error {
	if !withdrawal.Value.Valid {
		return fmt.Errorf("invalid value for withdrawal %d", withdrawal.ID)
	} else if !withdrawal.TradeFees.Valid {
		return fmt.Errorf("invalid tradeFees for withdrawal %d", withdrawal.ID)
	} else if !withdrawal.Fees.Valid {
		return fmt.Errorf("invalid fees for withdrawal %d", withdrawal.ID)
	}

	outputKey := withdrawal.CoinID
	withdrawalValue := withdrawal.Value.BigInt
	withdrawalFees := withdrawal.Fees.BigInt

	outputs, ok := m.outputs[outputKey]
	if !ok {
		return fmt.Errorf("unable to find output %s", outputKey)
	}

	totalOutputValue := new(big.Int)
	for _, entry := range outputs {
		totalOutputValue.Add(totalOutputValue, entry.actualOutput)
	}

	calculatedOutputValue := new(big.Int).Add(withdrawalValue, withdrawalFees)
	if calculatedOutputValue.Cmp(totalOutputValue) != 0 {
		return fmt.Errorf("output mismatch on withdrawal %d: have %s, want %s",
			withdrawal.ID, totalOutputValue, calculatedOutputValue)
	}

	usedValue := new(big.Int)
	for _, entry := range outputs {
		entry.finalOutput = adjustValue(withdrawalValue, entry.actualOutput, totalOutputValue)
		usedValue.Add(usedValue, entry.finalOutput)
	}

	// distribute remainders
	valueRemainder := new(big.Int).Sub(withdrawalValue, usedValue)
	if valueRemainder.Cmp(new(big.Int)) < 0 {
		return fmt.Errorf("negative remainder on withdrawal %d: have %s, want %s",
			withdrawal.ID, usedValue, withdrawalValue)
	} else {
		// since outputs is a map order can change, sort the keys
		// to make sure remainder distribution is deterministic
		keys := make([]string, 0)
		for key := range outputs {
			keys = append(keys, key)
		}

		sort.Strings(keys)
		if len(keys) == 0 {
			return fmt.Errorf("no entries to distribute the output remainder to")
		} else {
			entry := outputs[keys[0]]
			entry.finalOutput.Add(entry.finalOutput, valueRemainder)
		}
	}

	for _, entry := range outputs {
		entry.withdrawalFees = new(big.Int).Sub(entry.actualOutput, entry.finalOutput)
		// adjust deposit fees
		entry.depositFees = adjustValue(entry.depositFees, entry.finalOutput, entry.inputValue)
	}

	return nil
}

func (m *bimap) DistributeToBalances() error {
	for _, inputs := range m.inputs {
		for _, entry := range inputs {
			exchangeFees := new(big.Int)
			exchangeFees.Add(exchangeFees, entry.depositFees)
			exchangeFees.Add(exchangeFees, entry.tradeFees)
			exchangeFees.Add(exchangeFees, entry.withdrawalFees)

			totalValue := new(big.Int).Add(entry.finalOutput, exchangeFees)

			usedValue := new(big.Int)
			usedExchangeFees := new(big.Int)
			totalInValue := new(big.Int)
			for _, balance := range entry.balances {
				finalValue := adjustValue(entry.finalOutput, balance.InValue.BigInt, entry.inputValue)
				exchangeFees := adjustValue(exchangeFees, balance.InValue.BigInt, entry.inputValue)
				poolFees := adjustValue(balance.PoolFees.BigInt, totalValue, entry.inputValue)

				balance.OutValue = db.NullBigInt{finalValue, true}
				balance.ExchangeFees = db.NullBigInt{exchangeFees, true}
				balance.PoolFees = db.NullBigInt{poolFees, true}

				totalInValue.Add(totalInValue, balance.InValue.BigInt)
				usedValue.Add(usedValue, finalValue)
				usedExchangeFees.Add(usedExchangeFees, exchangeFees)
			}

			// credit remainder
			valueRemainder := new(big.Int).Sub(entry.finalOutput, usedValue)
			if valueRemainder.Cmp(new(big.Int)) < 0 {
				return fmt.Errorf("value remainder of %s is negative", valueRemainder)
			}

			exchangeFeeRemainder := new(big.Int).Sub(exchangeFees, usedExchangeFees)
			if exchangeFeeRemainder.Cmp(new(big.Int)) < 0 {
				return fmt.Errorf("exchange fee remainder of %s is negative", exchangeFeeRemainder)
			}

			for _, balance := range entry.balances {
				balance.OutValue.BigInt.Add(balance.OutValue.BigInt, valueRemainder)
				balance.ExchangeFees.BigInt.Add(balance.ExchangeFees.BigInt, exchangeFeeRemainder)
				break
			}
		}
	}

	return nil
}

func (m *bimap) DeleteByInput(k string) {
	if _, ok := m.inputs[k]; ok {
		delete(m.inputs, k)
	}

	for k2 := range m.outputs {
		for k3 := range m.outputs[k2] {
			if k3 == k {
				delete(m.outputs[k2], k3)
			}
		}
	}
}

func (m *bimap) DeleteByOutput(k string) {
	if _, ok := m.outputs[k]; ok {
		delete(m.outputs, k)
	}

	for k2 := range m.inputs {
		for k3 := range m.inputs[k2] {
			if k3 == k {
				delete(m.inputs[k2], k3)
			}
		}
	}
}

/* general reads */

func (m *bimap) GetByInput(inputKey, outputKey string) (*bimapEntry, bool) {
	if _, ok := m.inputs[inputKey]; !ok {
		return nil, false
	} else if _, ok := m.inputs[inputKey][outputKey]; !ok {
		return nil, false
	}

	return m.inputs[inputKey][outputKey], true
}

func (m *bimap) GetByOutput(outputKey, inputKey string) (*bimapEntry, bool) {
	if _, ok := m.outputs[outputKey]; !ok {
		return nil, false
	} else if _, ok := m.outputs[outputKey][inputKey]; !ok {
		return nil, false
	}

	return m.outputs[outputKey][inputKey], true
}

func (m *bimap) GetInputSum(inputKey string) (*big.Int, bool) {
	outputs, ok := m.inputs[inputKey]
	if !ok {
		return nil, false
	}

	value := new(big.Int)
	for _, entry := range outputs {
		if entry.inputValue != nil {
			value.Add(value, entry.inputValue)
		}
	}

	return value, true
}

func (m *bimap) GetOutputSum(outputKey string) (*big.Int, bool) {
	inputs, ok := m.outputs[outputKey]
	if !ok {
		return nil, false
	}

	value := new(big.Int)
	for _, entry := range inputs {
		if entry.estimateOutput != nil {
			value.Add(value, entry.estimateOutput)
		}
	}

	return value, true
}

func (m *bimap) GetOutputTradeSum(outputKey string) (*big.Int, bool) {
	inputs, ok := m.outputs[outputKey]
	if !ok {
		return nil, false
	}

	value := new(big.Int)
	for _, entry := range inputs {
		if entry.tradeValue != nil {
			value.Add(value, entry.tradeValue)
		}
	}

	return value, true
}
