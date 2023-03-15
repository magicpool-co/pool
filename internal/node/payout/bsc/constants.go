package bsc

import (
	"github.com/magicpool-co/pool/types"
)

func (node Node) Name() string {
	return "Binance Smart Chain"
}

func (node Node) Chain() string {
	return "BSC"
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

func (node Node) GetUnits() *types.Number {
	return new(types.Number).SetFromValue(1e18)
}

func (node Node) ValidateAddress(address string) bool {
	return true
}
