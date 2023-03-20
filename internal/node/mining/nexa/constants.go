package nexa

import (
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

const (
	addressCharset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

	pubKeyAddrID     = 0x00
	scriptHashAddrID = 0x08
	templateAddrID   = 0x98

	diffFactor = 4294967296
)

var (
	maxDiffBig   = common.MustParseBigHex("100000000000000000000000000000000000000000000000000000000")
	shareDiffBig = common.MustParseBigHex("500000000000000000000000000000000000000000000000000000000") // 0.2
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e8)
)

func (node Node) Name() string {
	return "Nexa"
}

func (node Node) Chain() string {
	return "NEXA"
}

func (node Node) Address() string {
	return node.address
}

func (node Node) GetAccountingType() types.AccountingType {
	return types.UTXOStructure
}

func (node Node) GetAddressPrefix() string {
	return "nexa"
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
	return 50
}

func (node Node) GetMatureDepth() uint64 {
	return 5000
}

func (node Node) CalculateHashrate(blockTime, difficulty float64) float64 {
	if blockTime == 0 || difficulty == 0 {
		return 0
	}
	return difficulty * (diffFactor / blockTime)
}

func ValidateAddress(address string) bool {
	return true
}

func (node Node) ValidateAddress(address string) bool {
	return ValidateAddress(address)
}
