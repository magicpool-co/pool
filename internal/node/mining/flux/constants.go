package flux

import (
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/types"
)

const (
	diffFactor = 8192
)

var (
	maxDiffBig   = common.MustParseBigHex("0007ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	shareDiffBig = common.MustParseBigHex("0007878787878787878787878787878787878787878787878787878787878786")
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e8)
)

func (node Node) Name() string {
	return "Flux"
}

func (node Node) Chain() string {
	return "FLUX"
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
	return 10
}

func (node Node) GetMatureDepth() uint64 {
	return 100
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
