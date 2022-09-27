package cfx

import (
	"math/big"

	cfxAddress "github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	maxDiffBig   = common.MustParseBigHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	shareDiffBig = common.MustParseBigHex("000000044b82fa09b5a52cb98b405447c4a98187eebb22f008d5d64f9c394ae8")
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e18)
)

func (node Node) Chain() string {
	return "CFX"
}

func (node Node) Address() string {
	return node.address
}

func (node Node) GetAccountingType() types.AccountingType {
	return types.AccountStructure
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
	return float64(shareDiff.Value())
}

func (node Node) GetMaxDifficulty() *big.Int {
	return maxDiffBig
}

func (node Node) GetImmatureDepth() uint64 {
	return 120
}

func (node Node) GetMatureDepth() uint64 {
	return 600
}

func (node Node) CalculateHashrate(blockTime, difficulty float64) float64 {
	if blockTime == 0 || difficulty == 0 {
		return 0
	}
	return difficulty / blockTime
}

func (node Node) ValidateAddress(address string) bool {
	parsedAddress, err := cfxAddress.NewFromBase32(address)
	if err != nil {
		return false
	} else if parsedAddress.GetNetworkID() != uint32(mainnetChainID) {
		return false
	}

	return parsedAddress.IsValid()
}
