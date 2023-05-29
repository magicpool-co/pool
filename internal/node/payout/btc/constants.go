package btc

import (
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) Name() string {
	return "Bitcoin"
}

func (node Node) Chain() string {
	return "BTC"
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

func (node Node) GetUnits() *types.Number {
	return new(types.Number).SetFromValue(1e8)
}

func (node Node) ShouldMergeUTXOs() bool {
	return false
}

func ValidateAddress(address string) bool {
	_, err := btctx.AddressToScript(address, mainnetPrefixP2PKH, mainnetPrefixP2SH, true)

	return err == nil
}

func (node Node) ValidateAddress(address string) bool {
	_, err := btctx.AddressToScript(address, node.prefixP2PKH, node.prefixP2SH, true)

	return err == nil
}
