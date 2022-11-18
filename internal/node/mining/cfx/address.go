package cfx

import (
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
)

const (
	typeBits    = 0
	addressType = 0
	sizeBits    = 0
	alphabet    = "abcdefghjkmnprstuvwxyz0123456789"
)

var (
	versionByte byte = (typeBits & 0x80) | (addressType << 3) | sizeBits
)

func ETHAddressToCFX(ethAddress []byte, networkPrefix string) (string, error) {
	return bech32.EncodeModified(alphabet, networkPrefix, versionByte, ethAddress)
}
