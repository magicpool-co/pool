package stats

import (
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/hako/durafmt"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
)

/* helpers */

func decimalsForChain(chain string) int {
	switch chain {
	case "NEXA":
		return 1
	case "KAS", "USD", "USDT", "BUSD":
		return 2
	case "BTC":
		return 6
	}

	return 4
}

type Number struct {
	Value     float64 `json:"value"`
	Formatted string  `json:"formatted"`
}

func newNumberFromUint64(value uint64) Number {
	n := Number{
		Value:     float64(value),
		Formatted: strconv.FormatUint(value, 10),
	}

	return n
}

func newNumberFromUint64Ptr(value uint64) *Number {
	n := newNumberFromUint64(value)
	return &n
}

func newNumberFromFloat64WithPrecision(value float64, precision int, units string, scaleUnits bool) Number {
	formattedValue := value
	if scaleUnits {
		scale, scaledValue := common.GetDefaultUnitScale(value)
		formattedValue /= scaledValue
		units = " " + scale + units
	}

	n := Number{
		Value:     value,
		Formatted: strconv.FormatFloat(formattedValue, 'f', precision, 64) + units,
	}

	return n
}

func newNumberFromFloat64(value float64, units string, scaleUnits bool) Number {
	return newNumberFromFloat64WithPrecision(value, 1, units, scaleUnits)
}

func newNumberFromFloat64ByChain(value float64, chain string) Number {
	decimals := decimalsForChain(chain)
	return newNumberFromFloat64WithPrecision(value, decimals, chain, false)
}

func newNumberFromFloat64Ptr(value float64, units string, scaleUnits bool) *Number {
	n := newNumberFromFloat64(value, units, scaleUnits)
	return &n
}

func newNumberFromBigInt(value *big.Int, chain string) (Number, error) {
	chain = strings.ToUpper(chain)
	units, err := common.GetDefaultUnits(chain)
	if err != nil {
		return Number{}, err
	}
	valueFloat := common.BigIntToFloat64(value, units)
	decimals := decimalsForChain(chain)

	n := Number{
		Value:     valueFloat,
		Formatted: strconv.FormatFloat(valueFloat, 'f', decimals, 64) + " " + chain,
	}

	return n, nil
}

func newNumberFromBigIntPtr(value *big.Int, chain string) (*Number, error) {
	n, err := newNumberFromBigInt(value, chain)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func newNumberFromDuration(value time.Duration) Number {
	duration, err := durafmt.ParseStringShort(value.String())
	if err != nil {
		return Number{}
	}

	n := Number{
		Value:     float64(value / time.Second),
		Formatted: duration.String(),
	}

	return n
}

/* pool */

type PoolSummary struct {
	Name               string `json:"name"`
	Symbol             string `json:"symbol"`
	Fee                Number `json:"fee"`
	Miners             uint64 `json:"miners"`
	Workers            uint64 `json:"workers"`
	Hashrate           Number `json:"hashrate"`
	Luck               Number `json:"luck"`
	TTF                Number `json:"ttf"`
	ProfitUSD          Number `json:"profitUsd"`
	ProfitBTC          Number `json:"profitBtc"`
	NetworkDifficulty  Number `json:"networkDifficulty"`
	NetworkHashrate    Number `json:"networkHashrate"`
	NetworkBlockReward Number `json:"networkBlockReward"`
}

/* dashboard */

type HashrateInfo struct {
	Hashrate    Number `json:"hashrate"`
	AvgHashrate Number `json:"avgHashrate"`
}

type ShareInfo struct {
	AcceptedShares    Number `json:"acceptedShares"`
	AcceptedShareRate Number `json:"acceptedShareRate"`
	RejectedShares    Number `json:"rejectedShares"`
	RejectedShareRate Number `json:"rejectedShareRate"`
	InvalidShares     Number `json:"invalidShares"`
	InvalidShareRate  Number `json:"invalidShareRate"`
}

type Miner struct {
	ID           uint64                   `json:"-"`
	Chain        string                   `json:"chain"`
	Address      string                   `json:"address"`
	Active       bool                     `json:"active"`
	HashrateInfo map[string]*HashrateInfo `json:"hashrateInfo"`
	ShareInfo    map[string]*ShareInfo    `json:"shareInfo"`
	FirstSeen    int64                    `json:"firstSeen"`
	LastSeen     int64                    `json:"lastSeen"`
}

type Worker struct {
	Name         string                   `json:"name"`
	Active       bool                     `json:"active"`
	HashrateInfo map[string]*HashrateInfo `json:"hashrateInfo"`
	ShareInfo    map[string]*ShareInfo    `json:"shareInfo"`
	FirstSeen    int64                    `json:"firstSeen"`
	LastSeen     int64                    `json:"lastSeen"`
}

type WorkerList struct {
	Active   []*Worker `json:"active"`
	Inactive []*Worker `json:"inactive"`
}

type Dashboard struct {
	Miners                  *Number                  `json:"miners,omitempty"`
	ActiveWorkers           *Number                  `json:"activeWorkers,omitempty"`
	InactiveWorkers         *Number                  `json:"inactiveWorkers,omitempty"`
	HashrateInfo            map[string]*HashrateInfo `json:"hashrateInfo"`
	ShareInfo               map[string]*ShareInfo    `json:"shareInfo"`
	ImmatureBalance         map[string]Number        `json:"immatureBalance"`
	PendingBalance          map[string]Number        `json:"pendingBalance"`
	UnpaidBalance           map[string]Number        `json:"unpaidBalance"`
	ProjectedEarningsNative map[string]Number        `json:"projectedEarningsNative"`
	ProjectedEarningsUSD    map[string]Number        `json:"projectedEarningsUsd"`
	ProjectedEarningsBTC    map[string]Number        `json:"projectedEarningsBtc"`
	ProjectedEarningsETH    map[string]Number        `json:"projectedEarningsEth"`
}

/* rounds */

type Round struct {
	Chain           string  `json:"chain"`
	Type            string  `json:"type"`
	Pending         bool    `json:"pending"`
	Mature          bool    `json:"mature"`
	Hash            string  `json:"hash"`
	Height          uint64  `json:"height"`
	ExplorerURL     string  `json:"explorerUrl"`
	Miner           *string `json:"miner"`
	Difficulty      Number  `json:"difficulty"`
	Hashrate        Number  `json:"hashrate"`
	Luck            Number  `json:"luck"`
	Value           Number  `json:"value"`
	MinerValue      *Number `json:"minerValue,omitempty"`
	MinerPercentage *Number `json:"minerPercentage,omitempty"`
	Timestamp       int64   `json:"timestamp"`
}

/* payouts */

type Payout struct {
	Chain        string `json:"chain"`
	Address      string `json:"address"`
	TxID         string `json:"txid"`
	ExplorerURL  string `json:"explorerUrl"`
	Confirmed    bool   `json:"confirmed"`
	Value        Number `json:"value"`
	PoolFees     Number `json:"poolFees"`
	ExchangeFees Number `json:"exchangeFees"`
	TxFees       Number `json:"txFees"`
	TotalFees    Number `json:"totalFees"`
	Timestamp    int64  `json:"timestamp"`
}

/* global chart */

type ChartSingle struct {
	Timestamps []int64              `json:"timestamps"`
	Values     map[string][]float64 `json:"values"`
}

/* block chart */

type BlockChart struct {
	Timestamp        []int64   `json:"timestamp"`
	Value            []float64 `json:"value,omitempty"`
	Difficulty       []float64 `json:"difficulty,omitempty"`
	BlockTime        []float64 `json:"blockTime,omitempty"`
	Hashrate         []float64 `json:"hashrate,omitempty"`
	UncleRate        []float64 `json:"uncleRate,omitempty"`
	Profitability    []float64 `json:"profitability,omitempty"`
	AvgProfitability []float64 `json:"avgProfitability,omitempty"`
	BlockCount       []uint64  `json:"blockCount,omitempty"`
	UncleCount       []uint64  `json:"uncleCount,omitempty"`
	TxCount          []uint64  `json:"txCount,omitempty"`
}

func (chart *BlockChart) Len() int {
	return len(chart.Timestamp)
}

func (chart *BlockChart) Swap(i, j int) {
	chart.Timestamp[i], chart.Timestamp[j] = chart.Timestamp[j], chart.Timestamp[i]
	chart.Value[i], chart.Value[j] = chart.Value[j], chart.Value[i]
	chart.Difficulty[i], chart.Difficulty[j] = chart.Difficulty[j], chart.Difficulty[i]
	chart.BlockTime[i], chart.BlockTime[j] = chart.BlockTime[j], chart.BlockTime[i]
	chart.Hashrate[i], chart.Hashrate[j] = chart.Hashrate[j], chart.Hashrate[i]
	chart.UncleRate[i], chart.UncleRate[j] = chart.UncleRate[j], chart.UncleRate[i]
	chart.Profitability[i], chart.Profitability[j] = chart.Profitability[j], chart.Profitability[i]
	chart.AvgProfitability[i], chart.AvgProfitability[j] = chart.AvgProfitability[j], chart.AvgProfitability[i]
	chart.BlockCount[i], chart.BlockCount[j] = chart.BlockCount[j], chart.BlockCount[i]
	chart.UncleCount[i], chart.UncleCount[j] = chart.UncleCount[j], chart.UncleCount[i]
	chart.TxCount[i], chart.TxCount[j] = chart.TxCount[j], chart.TxCount[i]
}

func (chart *BlockChart) Less(i, j int) bool {
	return chart.Timestamp[i] < chart.Timestamp[j]
}

func (chart *BlockChart) AddPoint(block *tsdb.Block) {
	chart.Timestamp = append(chart.Timestamp, block.EndTime.Unix())
	chart.Value = append(chart.Value, common.SafeRoundedFloat(block.Value, 3))
	chart.Difficulty = append(chart.Difficulty, common.SafeRoundedFloat(block.Difficulty, 3))
	chart.BlockTime = append(chart.BlockTime, common.SafeRoundedFloat(block.BlockTime, 3))
	chart.Hashrate = append(chart.Hashrate, common.SafeRoundedFloat(block.Hashrate, 3))
	chart.UncleRate = append(chart.UncleRate, common.SafeRoundedFloat(block.UncleRate, 5))
	chart.Profitability = append(chart.Profitability, block.Profitability)
	chart.AvgProfitability = append(chart.AvgProfitability, block.AvgProfitability)
	chart.BlockCount = append(chart.BlockCount, block.Count)
	chart.UncleCount = append(chart.UncleCount, block.UncleCount)
	chart.TxCount = append(chart.TxCount, block.TxCount)
}

/* round chart */

type RoundChart struct {
	Timestamp        []int64   `json:"timestamp"`
	Value            []float64 `json:"value"`
	Difficulty       []float64 `json:"difficulty"`
	RoundTime        []float64 `json:"roundTime"`
	Hashrate         []float64 `json:"hashrate"`
	UncleRate        []float64 `json:"uncleRate"`
	Luck             []float64 `json:"luck"`
	AvgLuck          []float64 `json:"avgLuck"`
	Profitability    []float64 `json:"profitability"`
	AvgProfitability []float64 `json:"avgProfitability"`
}

func (chart *RoundChart) Len() int {
	return len(chart.Timestamp)
}

func (chart *RoundChart) Swap(i, j int) {
	chart.Timestamp[i], chart.Timestamp[j] = chart.Timestamp[j], chart.Timestamp[i]
	chart.Value[i], chart.Value[j] = chart.Value[j], chart.Value[i]
	chart.Difficulty[i], chart.Difficulty[j] = chart.Difficulty[j], chart.Difficulty[i]
	chart.RoundTime[i], chart.RoundTime[j] = chart.RoundTime[j], chart.RoundTime[i]
	chart.Hashrate[i], chart.Hashrate[j] = chart.Hashrate[j], chart.Hashrate[i]
	chart.UncleRate[i], chart.UncleRate[j] = chart.UncleRate[j], chart.UncleRate[i]
	chart.Luck[i], chart.Luck[j] = chart.Luck[j], chart.Luck[i]
	chart.AvgLuck[i], chart.AvgLuck[j] = chart.AvgLuck[j], chart.AvgLuck[i]
	chart.Profitability[i], chart.Profitability[j] = chart.Profitability[j], chart.Profitability[i]
	chart.AvgProfitability[i], chart.AvgProfitability[j] = chart.AvgProfitability[j], chart.AvgProfitability[i]
}

func (chart *RoundChart) Less(i, j int) bool {
	return chart.Timestamp[i] < chart.Timestamp[j]
}

func (chart *RoundChart) AddPoint(round *tsdb.Round) {
	chart.Timestamp = append(chart.Timestamp, round.EndTime.Unix())
	chart.Value = append(chart.Value, common.SafeRoundedFloat(round.Value, 3))
	chart.Difficulty = append(chart.Difficulty, common.SafeRoundedFloat(round.Difficulty, 3))
	chart.RoundTime = append(chart.RoundTime, common.SafeRoundedFloat(round.RoundTime, 3))
	chart.Hashrate = append(chart.Hashrate, common.SafeRoundedFloat(round.Hashrate, 3))
	chart.UncleRate = append(chart.UncleRate, common.SafeRoundedFloat(round.UncleRate, 3))
	chart.Luck = append(chart.Luck, common.SafeRoundedFloat(round.Luck, 3))
	chart.AvgLuck = append(chart.AvgLuck, common.SafeRoundedFloat(round.AvgLuck, 3))
	chart.Profitability = append(chart.Profitability, common.SafeRoundedFloat(round.Profitability, 3))
	chart.AvgProfitability = append(chart.AvgProfitability, common.SafeRoundedFloat(round.AvgProfitability, 3))
}

/* share chart */

type ShareChart struct {
	Timestamp      []int64   `json:"timestamp"`
	Miners         []uint64  `json:"miners"`
	Workers        []uint64  `json:"workers"`
	AcceptedShares []uint64  `json:"acceptedShares"`
	RejectedShares []uint64  `json:"rejectedShares"`
	InvalidShares  []uint64  `json:"invalidShares"`
	Hashrate       []float64 `json:"hashrate"`
	AvgHashrate    []float64 `json:"avgHashrate"`
}

func (chart *ShareChart) Len() int {
	return len(chart.Timestamp)
}

func (chart *ShareChart) Swap(i, j int) {
	chart.Timestamp[i], chart.Timestamp[j] = chart.Timestamp[j], chart.Timestamp[i]
	chart.Miners[i], chart.Miners[j] = chart.Miners[j], chart.Miners[i]
	chart.Workers[i], chart.Workers[j] = chart.Workers[j], chart.Workers[i]
	chart.AcceptedShares[i], chart.AcceptedShares[j] = chart.AcceptedShares[j], chart.AcceptedShares[i]
	chart.RejectedShares[i], chart.RejectedShares[j] = chart.RejectedShares[j], chart.RejectedShares[i]
	chart.InvalidShares[i], chart.InvalidShares[j] = chart.InvalidShares[j], chart.InvalidShares[i]
	chart.Hashrate[i], chart.Hashrate[j] = chart.Hashrate[j], chart.Hashrate[i]
	chart.AvgHashrate[i], chart.AvgHashrate[j] = chart.AvgHashrate[j], chart.AvgHashrate[i]
}

func (chart *ShareChart) Less(i, j int) bool {
	return chart.Timestamp[i] < chart.Timestamp[j]
}

func (chart *ShareChart) AddPoint(share *tsdb.Share) {
	chart.Timestamp = append(chart.Timestamp, share.EndTime.Unix())
	chart.Miners = append(chart.Miners, share.Miners)
	chart.Workers = append(chart.Workers, share.Workers)
	chart.AcceptedShares = append(chart.AcceptedShares, share.AcceptedShares)
	chart.RejectedShares = append(chart.RejectedShares, share.RejectedShares)
	chart.InvalidShares = append(chart.InvalidShares, share.InvalidShares)
	chart.Hashrate = append(chart.Hashrate, common.SafeRoundedFloat(share.Hashrate, 3))
	chart.AvgHashrate = append(chart.AvgHashrate, common.SafeRoundedFloat(share.AvgHashrate, 3))
}
