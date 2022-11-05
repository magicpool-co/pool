package ergo

import (
	"bytes"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
	"github.com/magicpool-co/pool/types"
)

var (
	maxDiffBig   = common.MustParseBigHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	shareDiffBig = common.MustParseBigHex("7e00000007e00000007e00000007e00000007e00000007e00000007e")
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e9)
)

func (node Node) Chain() string {
	return "ERGO"
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
	return 72
}

func (node Node) GetMatureDepth() uint64 {
	return 720
}

func (node Node) CalculateHashrate(blockTime, difficulty float64) float64 {
	if blockTime == 0 || difficulty == 0 {
		return 0
	}
	return difficulty / blockTime
}

func ValidateAddress(address string) bool {
	const checksumLength = 4
	const minAddressLength = checksumLength + 2

	partial, err := base58.Decode(address)
	if err != nil {
		// address not base58
		return false
	} else if len(partial) < minAddressLength {
		// address too short
		return false
	} else if partial[0] >= 0x10 {
		// prefix not mainnet
		return false
	}

	// partial address without checksum
	withoutChecksum := partial[:len(partial)-checksumLength]

	// find given checksum, calculate actual checksum, check that they match
	givenChecksum := partial[len(partial)-checksumLength:]
	calculatedChecksum := crypto.Blake2b256(withoutChecksum)[:checksumLength]
	if bytes.Compare(givenChecksum, calculatedChecksum) != 0 {
		return false
	}

	// @TODO: properly parse the addresses like in
	// https://github.com/ergoplatform/sigma-rust/blob/481748de7f7b830dfeefe9d6433947e2e04a1179/ergotree-ir/src/chain/address.rs#L479-L488

	// check the address prefix
	switch partial[0] & 0xF {
	case 0x01: // Pay-to-PublicKey(P2PK)
	case 0x02: // Pay-to-Script-Hash(P2SH)
	case 0x03: // Pay-to-Script(P2S)
	default:
		return false
	}

	return true
}

func (node Node) ValidateAddress(address string) bool {
	return ValidateAddress(address)
}
