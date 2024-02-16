package mexc

import (
	"math/big"

	"github.com/magicpool-co/pool/types"
)

var (
	/* primary markets */
	cfxUsdtSell = &types.Market{Market: "CFXUSDT", Base: "CFX", Quote: "USDT", Direction: types.TradeSell}

	etcBtcSell = &types.Market{Market: "ETCUSDT", Base: "ETC", Quote: "USDT", Direction: types.TradeSell}

	firoUsdtSell = &types.Market{Market: "FIROUSDT", Base: "FIRO", Quote: "USDT", Direction: types.TradeSell}

	fluxUsdtSell = &types.Market{Market: "FLUX1USDT", Base: "FLUX1", Quote: "USDT", Direction: types.TradeSell}

	kasUsdtSell = &types.Market{Market: "KASUSDT", Base: "KAS", Quote: "USDT", Direction: types.TradeSell}

	nexaUsdtSell = &types.Market{Market: "NEXAUSDT", Base: "NEXA", Quote: "USDT", Direction: types.TradeSell}

	rvnUsdtSell = &types.Market{Market: "RVNUSDT", Base: "RVN", Quote: "USDT", Direction: types.TradeSell}

	/* secondary markets */
	busdUsdtBuy = &types.Market{Market: "BUSDUSDT", Base: "BUSD", Quote: "USDT", Direction: types.TradeBuy}
	btcUsdtBuy  = &types.Market{Market: "BTCUSDT", Base: "BTC", Quote: "USDT", Direction: types.TradeBuy}
	ethUsdtBuy  = &types.Market{Market: "ETHUSDT", Base: "ETH", Quote: "USDT", Direction: types.TradeBuy}
	usdcUsdtBuy = &types.Market{Market: "USDCUSDT", Base: "USDC", Quote: "USDT", Direction: types.TradeBuy}

	/* tertiary markets */
	btcBusdBuy = &types.Market{Market: "BTCBUSD", Base: "BTC", Quote: "BUSD", Direction: types.TradeBuy}
	ethBusdBuy = &types.Market{Market: "ETHBUSD", Base: "ETH", Quote: "BUSD", Direction: types.TradeBuy}
)

var (
	presetTradePaths = map[string]map[string][]*types.Market{
		"CFX": map[string][]*types.Market{
			"BTC": []*types.Market{
				cfxUsdtSell, busdUsdtBuy, btcBusdBuy,
			},
			"ETH": []*types.Market{
				cfxUsdtSell, busdUsdtBuy, ethBusdBuy,
			},
			"USDC": []*types.Market{
				cfxUsdtSell, usdcUsdtBuy,
			},
		},
		"ETC": map[string][]*types.Market{
			"BTC": []*types.Market{
				etcBtcSell, busdUsdtBuy, btcBusdBuy,
			},
			"ETH": []*types.Market{
				etcBtcSell, busdUsdtBuy, ethBusdBuy,
			},
			"USDC": []*types.Market{
				etcBtcSell, usdcUsdtBuy,
			},
		},
		"FIRO": map[string][]*types.Market{
			"BTC": []*types.Market{
				firoUsdtSell, busdUsdtBuy, btcBusdBuy,
			},
			"ETH": []*types.Market{
				firoUsdtSell, busdUsdtBuy, ethBusdBuy,
			},
			"USDC": []*types.Market{
				firoUsdtSell, usdcUsdtBuy,
			},
		},
		"FLUX": map[string][]*types.Market{
			"BTC": []*types.Market{
				fluxUsdtSell, busdUsdtBuy, btcBusdBuy,
			},
			"ETH": []*types.Market{
				fluxUsdtSell, busdUsdtBuy, ethBusdBuy,
			},
			"USDC": []*types.Market{
				fluxUsdtSell, usdcUsdtBuy,
			},
		},
		"KAS": map[string][]*types.Market{
			"BTC": []*types.Market{
				kasUsdtSell, busdUsdtBuy, btcBusdBuy,
			},
			"ETH": []*types.Market{
				kasUsdtSell, busdUsdtBuy, ethBusdBuy,
			},
			"USDC": []*types.Market{
				kasUsdtSell, usdcUsdtBuy,
			},
		},
		"NEXA": map[string][]*types.Market{
			"BTC": []*types.Market{
				nexaUsdtSell, busdUsdtBuy, btcBusdBuy,
			},
			"ETH": []*types.Market{
				nexaUsdtSell, busdUsdtBuy, ethBusdBuy,
			},
			"USDC": []*types.Market{
				nexaUsdtSell, usdcUsdtBuy,
			},
		},
		"RVN": map[string][]*types.Market{
			"BTC": []*types.Market{
				rvnUsdtSell, busdUsdtBuy, btcBusdBuy,
			},
			"ETH": []*types.Market{
				rvnUsdtSell, busdUsdtBuy, ethBusdBuy,
			},
			"USDC": []*types.Market{
				rvnUsdtSell, usdcUsdtBuy,
			},
		},
	}

	presetOutputThresholds = map[string]*big.Int{
		"BTC":  new(big.Int).SetUint64(2_500_000),               // 0.025 BTC
		"ETH":  new(big.Int).SetUint64(250_000_000_000_000_000), // 0.25 ETH
		"USDC": new(big.Int).SetUint64(20_000_000_000),          // 20,000 USDC
	}
)
