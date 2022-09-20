package eth

import (
	"regexp"

	"github.com/magicpool-co/pool/types"
)

var (
	units = new(types.Number).SetFromValue(1e18)

	addressExpr = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
)

func (node Node) Chain() string {
	if node.erc20 != nil {
		return node.erc20.Chain
	}

	return "ETH"
}

func (node Node) Address() string {
	return node.address
}

func (node Node) Mocked() bool {
	return node.mocked
}

func (node Node) GetUnits() *types.Number {
	if node.erc20 != nil {
		return node.erc20.Units
	}

	return units
}

func (node Node) ValidateAddress(address string) bool {
	return addressExpr.MatchString(address)
}
