package firo

import (
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/types"
)

const (
	diffFactor  = 4294967296
	epochLength = 1300
)

var (
	maxDiffBig   = common.MustParseBigHex("00000000ffff0000000000000000000000000000000000000000000000000000")
	shareDiffBig = common.MustParseBigHex("00000000ffff0000000000000000000000000000000000000000000000000000") // 1
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e8)
)

func (node Node) Name() string {
	return "Firo"
}

func (node Node) Chain() string {
	return "FIRO"
}

func (node Node) Address() string {
	return node.address
}

func (node Node) GetAccountingType() types.AccountingType {
	return types.UTXOStructure
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
	return float64(shareDiff.Value()) * diffFactor
}

func (node Node) GetMaxDifficulty() *big.Int {
	return maxDiffBig
}

func (node Node) GetImmatureDepth() uint64 {
	return 10
}

func (node Node) GetMatureDepth() uint64 {
	return 100
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
	_, err := btctx.AddressToScript(address, mainnetPrefixP2PKH, mainnetPrefixP2SH, false)

	return err == nil
}

func (node Node) ValidateAddress(address string) bool {
	_, err := btctx.AddressToScript(address, node.prefixP2PKH, node.prefixP2SH, false)

	return err == nil
}
