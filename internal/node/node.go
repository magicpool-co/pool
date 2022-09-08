package node

import (
	"fmt"
	"strings"

	"github.com/magicpool-co/pool/internal/node/mining/ae"
	"github.com/magicpool-co/pool/internal/node/mining/cfx"
	"github.com/magicpool-co/pool/internal/node/mining/ctxc"
	"github.com/magicpool-co/pool/internal/node/mining/ergo"
	"github.com/magicpool-co/pool/internal/node/mining/etc"
	"github.com/magicpool-co/pool/internal/node/mining/firo"
	"github.com/magicpool-co/pool/internal/node/mining/flux"
	"github.com/magicpool-co/pool/internal/node/mining/rvn"
	"github.com/magicpool-co/pool/internal/node/payout/bsc"
	"github.com/magicpool-co/pool/internal/node/payout/btc"
	"github.com/magicpool-co/pool/internal/node/payout/eth"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/types"
)

var (
	ErrUnsupportedChain = fmt.Errorf("unsupported chain")
)

func GetMiningNode(mainnet bool, chain, privKey string, urls []string, tunnel *sshtunnel.SSHTunnel) (types.MiningNode, error) {
	switch strings.ToUpper(chain) {
	case "AE":
		return ae.New(mainnet, urls, privKey, tunnel)
	case "CFX":
		return cfx.New(mainnet, urls, privKey, tunnel)
	case "CTXC":
		return ctxc.New(mainnet, urls, privKey, tunnel)
	case "ERGO":
		return ergo.New(mainnet, urls, privKey, tunnel)
	case "ETC":
		return etc.New(mainnet, urls, privKey, tunnel)
	case "FIRO":
		return firo.New(mainnet, urls, privKey, tunnel)
	case "FLUX":
		return flux.New(mainnet, urls, privKey, tunnel)
	case "RVN":
		return rvn.New(mainnet, urls, privKey, tunnel)
	default:
		return nil, ErrUnsupportedChain
	}
}

func GetPayoutNode(mainnet bool, chain, privKey, apiKey string, urls []string, tunnel *sshtunnel.SSHTunnel) (types.PayoutNode, error) {
	node, err := GetMiningNode(mainnet, chain, privKey, urls, tunnel)
	if err != nil && err != ErrUnsupportedChain {
		return nil, err
	} else if node != nil {
		return node, nil
	}

	switch strings.ToUpper(chain) {
	case "BSC":
		return bsc.New(mainnet, urls, privKey, tunnel)
	case "BTC":
		return btc.New(mainnet, privKey, apiKey)
	case "ETH":
		return eth.New(mainnet, urls, privKey, tunnel, nil)
	case "USDC":
		usdc := &eth.ERC20{
			Chain:    "USDC",
			Address:  "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
			Decimals: 6,
			Units:    new(types.Number).SetFromValue(1000000),
		}

		return eth.New(mainnet, urls, privKey, tunnel, usdc)
	default:
		return nil, ErrUnsupportedChain
	}
}
