package bittrex

import (
	"math/big"

	"github.com/magicpool-co/pool/types"
)

var (
	/* primary markets */
	etcBtcSell = &types.Market{Market: "ETC-BTC", Base: "ETC", Quote: "BTC", Direction: types.TradeSell}
	etcEthSell = &types.Market{Market: "ETC-ETH", Base: "ETC", Quote: "ETH", Direction: types.TradeSell}

	rvnBtcSell = &types.Market{Market: "RVN-BTC", Base: "RVN", Quote: "BTC", Direction: types.TradeSell}

	/* secondary markets */
	ethBtcBuy  = &types.Market{Market: "ETH-BTC", Base: "ETH", Quote: "BTC", Direction: types.TradeBuy}
	usdcBtcBuy = &types.Market{Market: "USDC-BTC", Base: "USDC", Quote: "BTC", Direction: types.TradeBuy}
	usdcEthBuy = &types.Market{Market: "USDC-ETH", Base: "USDC", Quote: "ETH", Direction: types.TradeBuy}
)

var (
	presetTradePaths = map[string]map[string][]*types.Market{
		"ETC": map[string][]*types.Market{
			"BTC": []*types.Market{
				etcBtcSell,
			},
			"ETH": []*types.Market{
				etcEthSell,
			},
			"USDC": []*types.Market{
				etcEthSell, usdcEthBuy,
			},
		},
		"RVN": map[string][]*types.Market{
			"BTC": []*types.Market{
				rvnBtcSell,
			},
			"ETH": []*types.Market{
				rvnBtcSell, ethBtcBuy,
			},
			"USDC": []*types.Market{
				rvnBtcSell, usdcBtcBuy,
			},
		},
	}

	presetOutputThresholds = map[string]*big.Int{
		"BTC":  new(big.Int).SetUint64(5_000_000),
		"ETH":  new(big.Int).SetUint64(500_000_000_000_000_000),
		"USDC": new(big.Int).SetUint64(2_000_000_000),
	}
)
