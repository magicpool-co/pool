package chart

import (
	"fmt"
	"sort"
	"time"

	"github.com/magicpool-co/pool/core/trade/binance"
	"github.com/magicpool-co/pool/core/trade/kucoin"
	"github.com/magicpool-co/pool/core/trade/mexc"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	pricePeriod = types.Period15m
	kucoinDelay = time.Second * 5
)

type priceIndex struct {
	usd map[time.Time]float64
	btc map[time.Time]float64
	eth map[time.Time]float64
}

func checkGaps(startTime, endTime time.Time, idx map[time.Time]float64) error {
	timeRange := make(map[int64]bool)
	for i := startTime; !i.After(endTime); i = i.Add(pricePeriod.Rollup()) {
		timeRange[i.Unix()] = true
	}

	for timestamp := range idx {
		delete(timeRange, timestamp.Unix())
	}

	gapList := make([]time.Time, 0)
	for timestamp := range timeRange {
		gapList = append(gapList, time.Unix(timestamp, 0))
	}

	sort.Slice(gapList, func(i, j int) bool {
		return gapList[i].Before(gapList[j])
	})

	var lastGap time.Time
	for i, timestamp := range gapList {
		if lastGap.IsZero() {
			lastGap = timestamp
		} else if timestamp.Sub(lastGap) != pricePeriod.Rollup() && i != len(gapList)-1 {
			return fmt.Errorf("range has an inconcievable gap %s -> %s", timestamp, lastGap)
		}
		lastGap = timestamp
	}

	if !lastGap.Equal(endTime) {
		return fmt.Errorf("gap is not at the end of the range")
	}

	return nil
}

func unifyRanges(a, b, c map[time.Time]float64) {
	toRemove := make(map[time.Time]bool)
	for timestamp := range a {
		if _, ok := b[timestamp]; !ok {
			toRemove[timestamp] = true
		} else if _, ok := c[timestamp]; !ok {
			toRemove[timestamp] = true
		}
	}

	for timestamp := range b {
		if _, ok := a[timestamp]; !ok {
			toRemove[timestamp] = true
		}
	}

	for timestamp := range c {
		if _, ok := a[timestamp]; !ok {
			toRemove[timestamp] = true
		}
	}

	for timestamp := range toRemove {
		delete(a, timestamp)
		delete(b, timestamp)
		delete(c, timestamp)
	}
}

func getKucoinNativeAll(chain string, startTime, endTime time.Time) (*priceIndex, error) {
	kucoinClient := kucoin.New("", "", "")
	usdt, err := kucoinClient.GetHistoricalRates(chain+"-USDT", startTime, endTime, false)
	if err != nil {
		return nil, err
	}
	time.Sleep(kucoinDelay)

	btc, err := kucoinClient.GetHistoricalRates(chain+"-BTC", startTime, endTime, false)
	if err != nil {
		return nil, err
	}
	time.Sleep(kucoinDelay)

	eth, err := kucoinClient.GetHistoricalRates(chain+"-ETH", startTime, endTime, false)
	if err != nil {
		return nil, err
	}

	unifyRanges(usdt, btc, eth)

	return &priceIndex{usd: usdt, btc: btc, eth: eth}, nil
}

func getKucoinNativeAllExceptETH(chain string, startTime, endTime time.Time) (*priceIndex, error) {
	kucoinClient := kucoin.New("", "", "")
	usdt, err := kucoinClient.GetHistoricalRates(chain+"-USDT", startTime, endTime, false)
	if err != nil {
		return nil, err
	}
	time.Sleep(kucoinDelay)

	btc, err := kucoinClient.GetHistoricalRates(chain+"-BTC", startTime, endTime, false)
	if err != nil {
		return nil, err
	}
	time.Sleep(kucoinDelay)

	eth, err := kucoinClient.GetHistoricalRates("ETH-USDT", startTime, endTime, true)
	if err != nil {
		return nil, err
	}

	unifyRanges(usdt, btc, eth)

	for timestamp, usdtPrice := range usdt {
		eth[timestamp] *= usdtPrice
	}

	return &priceIndex{usd: usdt, btc: btc, eth: eth}, nil
}

func getKucoinNativeUSDTOnly(chain string, startTime, endTime time.Time) (*priceIndex, error) {
	kucoinClient := kucoin.New("", "", "")
	usdt, err := kucoinClient.GetHistoricalRates(chain+"-USDT", startTime, endTime, false)
	if err != nil {
		return nil, err
	}
	time.Sleep(kucoinDelay)

	btc, err := kucoinClient.GetHistoricalRates("BTC-USDT", startTime, endTime, true)
	if err != nil {
		return nil, err
	}
	time.Sleep(kucoinDelay)

	eth, err := kucoinClient.GetHistoricalRates("ETH-USDT", startTime, endTime, true)
	if err != nil {
		return nil, err
	}

	unifyRanges(usdt, btc, eth)

	for timestamp, usdtPrice := range usdt {
		btc[timestamp] *= usdtPrice
		eth[timestamp] *= usdtPrice
	}

	return &priceIndex{usd: usdt, btc: btc, eth: eth}, nil
}

func getBinanceNativeAllExceptETH(chain string, startTime, endTime time.Time) (*priceIndex, error) {
	binanceClient := binance.New("", "")
	usdt, err := binanceClient.GetHistoricalRates(chain+"USDT", startTime, endTime, false)
	if err != nil {
		return nil, err
	}

	btc, err := binanceClient.GetHistoricalRates(chain+"BTC", startTime, endTime, false)
	if err != nil {
		return nil, err
	}

	eth, err := binanceClient.GetHistoricalRates("ETHUSDT", startTime, endTime, true)
	if err != nil {
		return nil, err
	}

	unifyRanges(usdt, btc, eth)

	for timestamp, usdtPrice := range usdt {
		eth[timestamp] *= usdtPrice
	}

	return &priceIndex{usd: usdt, btc: btc, eth: eth}, nil
}

func getMEXCNativeUSDTOnly(chain string, startTime, endTime time.Time) (*priceIndex, error) {
	mexcClient := mexc.New("", "")
	usdt, err := mexcClient.GetHistoricalRates(chain+"_USDT", startTime, endTime, false)
	if err != nil {
		return nil, err
	}

	btc, err := mexcClient.GetHistoricalRates("BTC_USDT", startTime, endTime, true)
	if err != nil {
		return nil, err
	}

	eth, err := mexcClient.GetHistoricalRates("ETH_USDT", startTime, endTime, true)
	if err != nil {
		return nil, err
	}

	unifyRanges(usdt, btc, eth)

	for timestamp, usdtPrice := range usdt {
		btc[timestamp] *= usdtPrice
		eth[timestamp] *= usdtPrice
	}

	return &priceIndex{usd: usdt, btc: btc, eth: eth}, nil
}

func (c *Client) ProcessPrices(chain string) error {
	startTime, err := tsdb.GetPriceMaxTimestamp(c.tsdb.Reader(), chain)
	if err != nil {
		return err
	} else if startTime.IsZero() {
		switch chain {
		case "ETHW":
			startTime = time.Unix(1664182800, 0)
		case "KAS":
			startTime = time.Unix(1664290800, 0)
		case "NEXA":
			startTime = time.Unix(1678147200, 0)
		default:
			startTime = time.Unix(1661990400, 0)
		}
	} else {
		startTime = startTime.Add(pricePeriod.Rollup())
	}

	endTime := common.NormalizeDate(time.Now().UTC(), pricePeriod.Rollup(), true)
	if !startTime.Add(time.Minute).Before(endTime) {
		return nil
	}

	var idx *priceIndex
	switch chain {
	case "CFX":
		idx, err = getBinanceNativeAllExceptETH("CFX", startTime, endTime)
	case "CTXC":
		idx, err = getBinanceNativeAllExceptETH("CTXC", startTime, endTime)
	case "ERGO":
		idx, err = getKucoinNativeAllExceptETH("ERG", startTime, endTime)
		if err == kucoin.ErrTooManyRequests {
			return nil
		}
	case "ETC":
		idx, err = getKucoinNativeUSDTOnly("ETC", startTime, endTime)
		if err == kucoin.ErrTooManyRequests {
			return nil
		}
	case "ETHW":
		idx, err = getKucoinNativeUSDTOnly("ETHW", startTime, endTime)
		if err == kucoin.ErrTooManyRequests {
			return nil
		}
	case "FIRO":
		idx, err = getBinanceNativeAllExceptETH("FIRO", startTime, endTime)
	case "FLUX":
		idx, err = getKucoinNativeAllExceptETH("FLUX", startTime, endTime)
		if err == kucoin.ErrTooManyRequests {
			return nil
		}
	case "KAS":
		idx, err = getMEXCNativeUSDTOnly("KAS", startTime, endTime)
		if err == mexc.ErrTooManyRequests {
			return nil
		}
	case "NEXA":
		idx, err = getMEXCNativeUSDTOnly("NEXA", startTime, endTime)
		if err == mexc.ErrTooManyRequests {
			return nil
		}
	case "RVN":
		idx, err = getKucoinNativeUSDTOnly("RVN", startTime, endTime)
		if err == kucoin.ErrTooManyRequests {
			return nil
		}
	}

	if err != nil || idx == nil {
		return err
	}

	var count int
	prices := make([]*tsdb.Price, len(idx.usd))
	for timestamp := range idx.usd {
		prices[count] = &tsdb.Price{
			ChainID:   chain,
			PriceUSD:  idx.usd[timestamp],
			PriceBTC:  idx.btc[timestamp],
			PriceETH:  idx.eth[timestamp],
			Timestamp: timestamp,
		}
		count++
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Timestamp.Before(prices[j].Timestamp)
	})

	return tsdb.InsertPrices(c.tsdb.Writer(), prices...)
}
