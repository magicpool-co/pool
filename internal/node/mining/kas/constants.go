package kas

import (
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	"github.com/magicpool-co/pool/types"
)

const (
	addressCharset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

	pubKeyAddrID      = 0x00
	pubKeyECDSAAddrID = 0x01
	scriptHashAddrID  = 0x08

	pubKeySize      = 32
	pubKeySizeECDSA = 33
)

var (
	// note: pools use btc diff (2^256 - 1), but native kaspa diff is actually half of btc diff (2^255 - 1).
	// this means, if using btc diff, hashrate = (2 * diff) / blocktime and kaspa diff = diff / 2.
	maxDiffBig   = common.MustParseBigHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	shareDiffBig = common.MustParseBigHex("0000000019998000000000b876f1ba8c727da45575f4fafc1a8cee77caf0878d")
	shareDiff    = new(types.Difficulty).SetFromBig(shareDiffBig, maxDiffBig)
	units        = new(types.Number).SetFromValue(1e8)
)

func (node Node) Chain() string {
	return "KAS"
}

func (node Node) Address() string {
	return node.address
}

func (node Node) GetAccountingType() types.AccountingType {
	return types.UTXOStructure
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
	return 50
}

func (node Node) GetMatureDepth() uint64 {
	return 500
}

func (node Node) CalculateHashrate(blockTime, difficulty float64) float64 {
	if blockTime == 0 || difficulty == 0 {
		return 0
	}
	return (4 * difficulty) / blockTime
}

func ValidateAddress(address string) bool {
	prefix, version, decoded, err := bech32.DecodeBCH(addressCharset, address)
	if err != nil {
		return false
	} else if prefix != mainnetPrefix {
		return false
	}

	switch version {
	case pubKeyAddrID, scriptHashAddrID:
		return len(decoded) == pubKeySize
	case pubKeyECDSAAddrID:
		return len(decoded) == pubKeySizeECDSA
	default:
		return false
	}
}

func (node Node) ValidateAddress(address string) bool {
	prefix, version, decoded, err := bech32.DecodeBCH(addressCharset, address)
	if err != nil {
		return false
	} else if prefix != node.prefix {
		return false
	}

	switch version {
	case pubKeyAddrID, scriptHashAddrID:
		return len(decoded) == pubKeySize
	case pubKeyECDSAAddrID:
		return len(decoded) == pubKeySizeECDSA
	default:
		return false
	}
}
