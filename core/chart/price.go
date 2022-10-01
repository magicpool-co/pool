package chart

import (
	"fmt"
	"sort"
	"time"

	"github.com/magicpool-co/pool/core/trade/binance"
	"github.com/magicpool-co/pool/core/trade/kucoin"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	priceStart  = time.Unix(1661990400, 0)
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

func (c *Client) ProcessPrices(chain string) error {
	startTime, err := tsdb.GetPriceMaxTimestamp(c.tsdb.Reader(), chain)
	if err != nil {
		return err
	} else if startTime.IsZero() {
		startTime = priceStart
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
	case "ETC":
		idx, err = getKucoinNativeAll("ETC", startTime, endTime)
	case "ETHW":
		idx, err = getKucoinNativeUSDTOnly("ETHW", startTime, endTime)
	case "FIRO":
		idx, err = getBinanceNativeAllExceptETH("FIRO", startTime, endTime)
	case "FLUX":
		idx, err = getKucoinNativeAllExceptETH("FLUX", startTime, endTime)
	case "RVN":
		idx, err = getKucoinNativeUSDTOnly("RVN", startTime, endTime)
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