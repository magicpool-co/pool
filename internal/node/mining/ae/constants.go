package ae

import (
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

const (
	diffFactor = 42

	addressPrefix = "ak_"
	txPrefix      = "tx_"
	txHashPrefix  = "th_"
)

var (
	maxDiffBig = common.MustParseBigHex("ffff000000000000000000000000000000000000000000000000000000000000")
	shareDiff  = new(types.Difficulty).SetFromValue(16, maxDiffBig)
	units      = new(types.Number).SetFromValue(1e18)
)

func (node Node) Name() string {
	return "Aeternity"
}

func (node Node) Chain() string {
	return "AE"
}

func (node Node) Address() string {
	return node.address
}

func (node Node) GetAccountingType() types.AccountingType {
	return types.AccountStructure
}

func (node Node) GetAddressPrefix() string {
	return ""
}

func (node Node) Mocked() bool {
	return node.mocked
}

func (node Node) GetUnits() *types.Number {
	return units
}

func (node Node) GetShareDifficulty() *types.Difficulty {
	return shareDiff
}

func (node Node) GetAdjustedShareDifficulty() float64 {
	return float64(shareDiff.Value()) * diffFactor
}

func (node Node) GetMaxDifficulty() *big.Int {
	return maxDiffBig
}

func (node Node) GetImmatureDepth() uint64 {
	return 18
}

func (node Node) GetMatureDepth() uint64 {
	return 180
}

func (node Node) CalculateHashrate(blockTime, difficulty float64) float64 {
	if blockTime == 0 || difficulty == 0 {
		return 0
	}
	return difficulty * (diffFactor / blockTime)
}

func (node Node) ValidateAddress(address string) bool {
	return true
}
