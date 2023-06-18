package rvn

import (
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/types"
)

const (
	globalDiffFactor = 4294967296
	epochLength      = 7500
)

var (
	maxDiffBig   = common.MustParseBigHex("00000000ffff0000000000000000000000000000000000000000000000000000")
	shareDiffBig = common.MustParseBigHex("00000000ffff0000000000000000000000000000000000000000000000000000") // 1
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e8)
)

func (node Node) Name() string {
	return "Ravencoin"
}

func (node Node) Chain() string {
	return "RVN"
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

func (node Node) GetShareDifficulty(diffFactor int) *types.Difficulty {
	if diffFactor > 1 {
		return shareDiff.Mul(int64(diffFactor))
	}
	return shareDiff
}

func (node Node) GetAdjustedShareDifficulty() float64 {
	return float64(shareDiff.Value()) * globalDiffFactor
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
	return difficulty * (globalDiffFactor / blockTime)
}

func ValidateAddress(address string) bool {
	_, err := btctx.AddressToScript(address, mainnetPrefixP2PKH, nil, false)

	return err == nil
}

func (node Node) ValidateAddress(address string) bool {
	return ValidateAddress(address)
}
