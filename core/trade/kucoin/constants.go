package kucoin

import (
	"math/big"

	"github.com/magicpool-co/pool/types"
)

var (
	/* primary markets */
	cfxUsdtSell = &types.Market{Market: "CFX-USDT", Base: "CFX", Quote: "USDT", Direction: types.TradeSell}

	ergBtcSell  = &types.Market{Market: "ERG-BTC", Base: "ERG", Quote: "BTC", Direction: types.TradeSell}
	ergUsdtSell = &types.Market{Market: "ERG-USDT", Base: "ERG", Quote: "USDT", Direction: types.TradeSell}

	etcBtcSell  = &types.Market{Market: "ETC-BTC", Base: "ETC", Quote: "BTC", Direction: types.TradeSell}
	etcEthSell  = &types.Market{Market: "ETC-ETH", Base: "ETC", Quote: "ETH", Direction: types.TradeSell}
	etcUsdcSell = &types.Market{Market: "ETC-USDC", Base: "ETC", Quote: "USDC", Direction: types.TradeSell}

	fluxBtcSell  = &types.Market{Market: "FLUX-BTC", Base: "FLUX", Quote: "BTC", Direction: types.TradeSell}
	fluxUsdtSell = &types.Market{Market: "FLUX-USDT", Base: "FLUX", Quote: "USDT", Direction: types.TradeSell}

	rvnUsdtSell = &types.Market{Market: "RVN-USDT", Base: "RVN", Quote: "USDT", Direction: types.TradeSell}

	/* secondary markets */
	btcUsdtBuy  = &types.Market{Market: "BTC-USDT", Base: "BTC", Quote: "USDT", Direction: types.TradeBuy}
	ethUsdtBuy  = &types.Market{Market: "ETH-USDT", Base: "ETH", Quote: "USDT", Direction: types.TradeBuy}
	usdcUsdtBuy = &types.Market{Market: "USDC-USDT", Base: "USDC", Quote: "USDT", Direction: types.TradeBuy}

	ethBtcBuy = &types.Market{Market: "ETH-BTC", Base: "ETH", Quote: "BTC", Direction: types.TradeBuy}
)

var (
	presetTradePaths = map[string]map[string][]*types.Market{
		"CFX": map[string][]*types.Market{
			"BTC": []*types.Market{
				cfxUsdtSell, btcUsdtBuy,
			},
			"ETH": []*types.Market{
				cfxUsdtSell, ethUsdtBuy,
			},
			"USDC": []*types.Market{
				cfxUsdtSell, usdcUsdtBuy,
			},
		},
		"ERG": map[string][]*types.Market{
			"BTC": []*types.Market{
				ergBtcSell,
			},
			"ETH": []*types.Market{
				ergBtcSell, ethBtcBuy,
			},
			"USDC": []*types.Market{
				ergUsdtSell, usdcUsdtBuy,
			},
		},
		"ETC": map[string][]*types.Market{
			"BTC": []*types.Market{
				etcBtcSell,
			},
			"ETH": []*types.Market{
				etcEthSell,
			},
			"USDC": []*types.Market{
				etcUsdcSell,
			},
		},
		"FLUX": map[string][]*types.Market{
			"BTC": []*types.Market{
				fluxBtcSell,
			},
			"ETH": []*types.Market{
				fluxBtcSell, ethBtcBuy,
			},
			"USDC": []*types.Market{
				fluxUsdtSell, usdcUsdtBuy,
			},
		},
		"RVN": map[string][]*types.Market{
			"BTC": []*types.Market{
				rvnUsdtSell, btcUsdtBuy,
			},
			"ETH": []*types.Market{
				rvnUsdtSell, ethUsdtBuy,
			},
			"USDC": []*types.Market{
				rvnUsdtSell, usdcUsdtBuy,
			},
		},
	}

	presetOutputThresholds = map[string]*big.Int{
		"BTC":  new(big.Int).SetUint64(5_000_000),
		"ETH":  new(big.Int).SetUint64(500_000_000_000_000_000),
		"USDC": new(big.Int).SetUint64(2_000_000_000),
	}
)
