package binance

import (
	"math/big"

	"github.com/magicpool-co/pool/types"
)

var (
	/* primary markets */
	cfxBtcSell  = &types.Market{Market: "CFXBTC", Base: "CFX", Quote: "BTC", Direction: types.TradeSell}
	cfxUsdtSell = &types.Market{Market: "CFXUSDT", Base: "CFX", Quote: "USDT", Direction: types.TradeSell}

	ctxcBtcSell  = &types.Market{Market: "CTXCBTC", Base: "CTXC", Quote: "BTC", Direction: types.TradeSell}
	ctxcUsdtSell = &types.Market{Market: "CTXCUSDT", Base: "CTXC", Quote: "USDT", Direction: types.TradeSell}

	etcBtcSell  = &types.Market{Market: "ETCBTC", Base: "ETC", Quote: "BTC", Direction: types.TradeSell}
	etcEthSell  = &types.Market{Market: "ETCETH", Base: "ETC", Quote: "ETH", Direction: types.TradeSell}
	etcUsdtSell = &types.Market{Market: "ETCUSDT", Base: "ETC", Quote: "USDT", Direction: types.TradeSell}

	firoBtcSell  = &types.Market{Market: "FIROBTC", Base: "FIRO", Quote: "BTC", Direction: types.TradeSell}
	firoUsdtSell = &types.Market{Market: "FIROUSDT", Base: "FIRO", Quote: "USDT", Direction: types.TradeSell}

	rvnBtcSell  = &types.Market{Market: "RVNBTC", Base: "RVN", Quote: "BTC", Direction: types.TradeSell}
	rvnUsdtSell = &types.Market{Market: "RVNUSDT", Base: "RVN", Quote: "USDT", Direction: types.TradeSell}

	/* secondary markets */
	btcUsdtBuy  = &types.Market{Market: "BTCUSDT", Base: "BTC", Quote: "USDT", Direction: types.TradeBuy}
	ethUsdtBuy  = &types.Market{Market: "ETHUSDT", Base: "ETH", Quote: "USDT", Direction: types.TradeBuy}
	usdcUsdtBuy = &types.Market{Market: "USDCUSDT", Base: "USDC", Quote: "USDT", Direction: types.TradeBuy}
)

var (
	presetTradePaths = map[string]map[string][]*types.Market{
		"CFX": map[string][]*types.Market{
			"BTC": []*types.Market{
				cfxBtcSell,
			},
			"ETH": []*types.Market{
				cfxUsdtSell, ethUsdtBuy,
			},
			"USDC": []*types.Market{
				cfxUsdtSell, usdcUsdtBuy,
			},
		},
		"CTXC": map[string][]*types.Market{
			"BTC": []*types.Market{
				ctxcBtcSell,
			},
			"ETH": []*types.Market{
				ctxcUsdtSell, ethUsdtBuy,
			},
			"USDC": []*types.Market{
				ctxcUsdtSell, usdcUsdtBuy,
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
				etcUsdtSell, usdcUsdtBuy,
			},
		},
		"FIRO": map[string][]*types.Market{
			"BTC": []*types.Market{
				firoBtcSell,
			},
			"ETH": []*types.Market{
				firoUsdtSell, ethUsdtBuy,
			},
			"USDC": []*types.Market{
				firoUsdtSell, usdcUsdtBuy,
			},
		},
		"RVN": map[string][]*types.Market{
			"BTC": []*types.Market{
				rvnBtcSell,
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
