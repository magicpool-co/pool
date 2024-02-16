package api

import (
	"strings"

	"github.com/magicpool-co/pool/internal/node/mining/cfx"
	"github.com/magicpool-co/pool/internal/node/mining/erg"
	"github.com/magicpool-co/pool/internal/node/mining/etc"
	"github.com/magicpool-co/pool/internal/node/mining/firo"
	"github.com/magicpool-co/pool/internal/node/mining/flux"
	"github.com/magicpool-co/pool/internal/node/mining/kas"
	"github.com/magicpool-co/pool/internal/node/mining/nexa"
	"github.com/magicpool-co/pool/internal/node/mining/rvn"
	"github.com/magicpool-co/pool/internal/node/payout/btc"
	"github.com/magicpool-co/pool/internal/node/payout/eth"
)

func validateMiningChain(chain string) bool {
	switch strings.ToUpper(chain) {
	case "CFX", "ERG", "ETC", "ETHW", "FIRO", "FLUX", "KAS", "NEXA", "RVN":
		return true
	default:
		return false
	}
}

func validatePayoutChain(chain string) bool {
	if validateMiningChain(chain) {
		return true
	}

	switch strings.ToUpper(chain) {
	case "BTC", "ETH", "USDC":
		return true
	default:
		return false
	}
}

func parseMiner(miner string) (string, string, error) {
	parts := strings.Split(miner, ":")
	if len(parts) != 2 {
		return "", "", errMinerNotFound
	}

	switch strings.ToLower(parts[0]) {
	case "cfx":
		parts[0] = "CFX"
		parts[1] = miner
	case "ergo":
		parts[0] = "ERG"
	case "kaspa":
		parts[0] = "KAS"
		parts[1] = miner
	case "nexa":
		parts[0] = "NEXA"
		parts[1] = miner
	}

	if !validatePayoutChain(parts[0]) {
		return "", "", errChainNotFound
	}

	return strings.ToUpper(parts[0]), parts[1], nil
}

func ValidateAddress(chain, address string) bool {
	switch strings.ToUpper(chain) {
	case "BTC":
		return btc.ValidateAddress(address)
	case "CFX":
		return cfx.ValidateAddress(address)
	case "ERG":
		return erg.ValidateAddress(address)
	case "ETC", "ETHW":
		return etc.ValidateAddress(address)
	case "BSC", "ETH", "USDC":
		return eth.ValidateAddress(address)
	case "FIRO":
		return firo.ValidateAddress(address)
	case "FLUX":
		return flux.ValidateAddress(address)
	case "KAS":
		return kas.ValidateAddress(address)
	case "NEXA":
		return nexa.ValidateAddress(address)
	case "RVN":
		return rvn.ValidateAddress(address)
	default:
		return false
	}
}
