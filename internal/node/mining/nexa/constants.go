package nexa

import (
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/nexatx"
	"github.com/magicpool-co/pool/types"
)

const (
	addressCharset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

	pubKeyAddrID     = 0x00
	scriptHashAddrID = 0x08
	templateAddrID   = 0x98

	diffFactor  = 4294967296
	shareFactor = 0.2
)

var (
	maxDiffBig   = common.MustParseBigHex("100000000000000000000000000000000000000000000000000000000")
	shareDiffBig = common.MustParseBigHex("500000000000000000000000000000000000000000000000000000000") // 0.2 (shareFactor)
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e2)
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

func (node Node) GetShareDifficulty(shareFactor int64) *types.Difficulty {
	if shareFactor > 1 {
		return shareDiff.Mul(shareFactor)
	}
	return shareDiff
}

func (node Node) GetAdjustedShareDifficulty() float64 {
	return diffFactor * shareFactor
}

func (node Node) GetMaxDifficulty() *big.Int {
	return maxDiffBig
}

func (node Node) GetImmatureDepth() uint64 {
	return 25
}

func (node Node) GetMatureDepth() uint64 {
	return 5000
}

func (node Node) ShouldMergeUTXOs() bool {
	return false
}

func (node Node) CalculateHashrate(blockTime, difficulty float64) float64 {
	if blockTime == 0 || difficulty == 0 {
		return 0
	}
	return difficulty * (diffFactor / blockTime)
}

func ValidateAddress(address string) bool {
	_, _, err := nexatx.AddressToScript(address, mainnetPrefix)

	return err == nil
}

func (node Node) ValidateAddress(address string) bool {
	return ValidateAddress(address)
}
