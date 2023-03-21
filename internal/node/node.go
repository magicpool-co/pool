package node

import (
	"fmt"
	"strings"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/node/mining/ae"
	"github.com/magicpool-co/pool/internal/node/mining/cfx"
	"github.com/magicpool-co/pool/internal/node/mining/ctxc"
	"github.com/magicpool-co/pool/internal/node/mining/erg"
	"github.com/magicpool-co/pool/internal/node/mining/etc"
	"github.com/magicpool-co/pool/internal/node/mining/firo"
	"github.com/magicpool-co/pool/internal/node/mining/flux"
	"github.com/magicpool-co/pool/internal/node/mining/kas"
	"github.com/magicpool-co/pool/internal/node/mining/nexa"
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

func GetMiningNode(mainnet bool, chain, privKey string, urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (types.MiningNode, error) {
	switch strings.ToUpper(chain) {
	case "AE":
		return ae.New(mainnet, urls, privKey, logger, tunnel)
	case "CFX":
		return cfx.New(mainnet, urls, privKey, logger, tunnel)
	case "CTXC":
		return ctxc.New(mainnet, urls, privKey, logger, tunnel)
	case "ERG":
		return erg.New(mainnet, urls, privKey, logger, tunnel)
	case "ETC":
		return etc.New(etc.ETC, mainnet, urls, privKey, logger, tunnel)
	case "ETHW":
		return etc.New(etc.ETHW, mainnet, urls, privKey, logger, tunnel)
	case "FIRO":
		return firo.New(mainnet, urls, privKey, logger, tunnel)
	case "FLUX":
		return flux.New(mainnet, urls, privKey, logger, tunnel)
	case "KAS":
		return kas.New(mainnet, urls, privKey, logger, tunnel)
	case "NEXA":
		return nexa.New(mainnet, urls, privKey, logger, tunnel)
	case "RVN":
		return rvn.New(mainnet, urls, privKey, logger, tunnel)
	default:
		return nil, ErrUnsupportedChain
	}
}

func GetPayoutNode(mainnet bool, chain, privKey, apiKey, url string, logger *log.Logger) (types.PayoutNode, error) {
	node, err := GetMiningNode(mainnet, chain, privKey, []string{url}, logger, nil)
	if err != nil && err != ErrUnsupportedChain {
		return nil, err
	} else if node != nil {
		return node, nil
	}

	switch strings.ToUpper(chain) {
	case "BSC":
		return bsc.New(mainnet, url, privKey, logger)
	case "BTC":
		return btc.New(mainnet, privKey, apiKey)
	case "ETH":
		return eth.New(mainnet, url, privKey, nil, logger)
	case "USDC":
		usdc := &eth.ERC20{
			Chain:    "USDC",
			Address:  "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
			Decimals: 6,
			Units:    new(types.Number).SetFromValue(1000000),
		}

		return eth.New(mainnet, url, privKey, usdc, logger)
	default:
		return nil, ErrUnsupportedChain
	}
}
