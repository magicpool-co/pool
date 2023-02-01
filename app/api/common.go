package api

import (
	"strings"

	"github.com/magicpool-co/pool/internal/node/mining/cfx"
	"github.com/magicpool-co/pool/internal/node/mining/ctxc"
	"github.com/magicpool-co/pool/internal/node/mining/ergo"
	"github.com/magicpool-co/pool/internal/node/mining/etc"
	"github.com/magicpool-co/pool/internal/node/mining/firo"
	"github.com/magicpool-co/pool/internal/node/mining/flux"
	"github.com/magicpool-co/pool/internal/node/mining/kas"
	"github.com/magicpool-co/pool/internal/node/mining/rvn"
	"github.com/magicpool-co/pool/internal/node/payout/btc"
	"github.com/magicpool-co/pool/internal/node/payout/eth"
)

func validateMiningChain(chain string) bool {
	switch strings.ToUpper(chain) {
	case "CFX", "CTXC", "ERGO", "ETC", "ETHW", "FIRO", "FLUX", "KAS", "RVN":
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
	case "kaspa":
		parts[0] = "KAS"
		parts[1] = miner
	case "cfx":
		parts[0] = "CFX"
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
		return cfx.ValidateAddress("cfx:" + address)
	case "CTXC":
		return ctxc.ValidateAddress(address)
	case "ERGO":
		return ergo.ValidateAddress(address)
	case "ETC", "ETHW":
		return etc.ValidateAddress(address)
	case "BSC", "ETH", "USDC":
		return eth.ValidateAddress(address)
	case "FIRO":
		return firo.ValidateAddress(address)
	case "FLUX":
		return flux.ValidateAddress(address)
	case "KAS":
		return kas.ValidateAddress("kaspa:" + address)
	case "RVN":
		return rvn.ValidateAddress(address)
	default:
		return false
	}
}
