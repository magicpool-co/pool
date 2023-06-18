package etc

import (
	"math/big"
	"regexp"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	maxDiffBig   = common.MustParseBigHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	shareDiffBig = common.MustParseBigHex("7e00000007e00000007e00000007e00000007e00000007e00000007e")
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e18)

	addressExpr = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
)

func (node Node) Name() string {
	switch node.ethType {
	case ETC:
		return "Ethereum Classic"
	case ETHW:
		return "EthereumPoW"
	default:
		return ""
	}
}

func (node Node) Chain() string {
	switch node.ethType {
	case ETC:
		return "ETC"
	case ETHW:
		return "ETHW"
	default:
		return ""
	}
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

func (node Node) GetShareDifficulty(shareFactor int64) *types.Difficulty {
	if shareFactor > 1 {
		return shareDiff.Mul(shareFactor)
	}
	return shareDiff
}

func (node Node) GetAdjustedShareDifficulty() float64 {
	return float64(shareDiff.Value())
}

func (node Node) GetMaxDifficulty() *big.Int {
	return maxDiffBig
}

func (node Node) GetImmatureDepth() uint64 {
	return 25
}

func (node Node) GetMatureDepth() uint64 {
	return 120
}

func (node Node) ShouldMergeUTXOs() bool {
	return false
}

func (node Node) CalculateHashrate(blockTime, difficulty float64) float64 {
	if blockTime == 0 || difficulty == 0 {
		return 0
	}
	return difficulty / blockTime
}

func ValidateAddress(address string) bool {
	return addressExpr.MatchString(address)
}

func (node Node) ValidateAddress(address string) bool {
	return ValidateAddress(address)
}
