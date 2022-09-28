package stats

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/magicpool-co/pool/pkg/common"
)

/* helpers */

type Number struct {
	Value     float64 `json:"value"`
	Formatted string  `json:"formatted"`
	Units     string  `json:"units"`
}

func newNumberFromFloat64(value float64, units string, scaleUnits bool) Number {
	if scaleUnits {
		scale, scaledValue := common.GetDefaultUnitScale(value)
		value /= scaledValue
		units = " " + units + scale
	}

	n := Number{
		Value:     value,
		Formatted: strconv.FormatFloat(value, 'f', 0, 64) + units,
		Units:     units,
	}

	return n
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

	n := Number{
		Value:     valueFloat,
		Formatted: strconv.FormatFloat(valueFloat, 'f', 4, 64) + " " + chain,
		Units:     " " + chain,
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

/* dashboard */

type HashrateInfo struct {
	Hashrate         Number `json:"hashrate"`
	AvgHashrate      Number `json:"avgHashrate"`
	ReportedHashrate Number `json:"reportedHashrate"`
}

type ShareInfo struct {
	AcceptedShares    Number `json:"acceptedShares"`
	AcceptedShareRate Number `json:"acceptedShareRate"`
	RejectedShares    Number `json:"rejectedShares"`
	RejectedShareRate Number `json:"rejectedShareRate"`
	InvalidShares     Number `json:"invalidShares"`
	InvalidShareRate  Number `json:"invalidShareRate"`
}

type Worker struct {
	Name           string `json:"name"`
	Active         bool   `json:"active"`
	Hashrate       Number `json:"hashrate"`
	AvgHashrate    Number `json:"avgHashrate"`
	AcceptedShares Number `json:"acceptedShares"`
	RejectedShares Number `json:"rejectedShares"`
	InvalidShares  Number `json:"invalidShares"`
	LastSeen       int64  `json:"lastSeen"`
}

type Dashboard struct {
	MinersCount     *Number                  `json:"minersCount,omitempty"`
	WorkersCount    *Number                  `json:"workersCount,omitempty"`
	WorkersActive   []*Worker                `json:"workersActive,omitempty"`
	WorkersInactive []*Worker                `json:"workersInactive,omitempty"`
	Hashrate        map[string]*HashrateInfo `json:"hashrate"`
	Shares          map[string]*ShareInfo    `json:"shareInfo"`
	PendingBalance  map[string]Number        `json:"pendingBalance"`
	UnpaidBalance   map[string]Number        `json:"unpaidBalance"`
}

/* blocks */

type Block struct {
	Chain           string  `json:"chain"`
	Type            string  `json:"type"`
	Pending         bool    `json:"pending"`
	Mature          bool    `json:"mature"`
	Height          uint64  `json:"height"`
	ExplorerURL     string  `json:"explorerUrl"`
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
	Value        Number `json:"value"`
	PoolFees     Number `json:"poolFees"`
	ExchangeFees Number `json:"exchangeFees"`
	TxFees       Number `json:"txFees"`
	TotalFees    Number `json:"totalFees"`
	Timestamp    int64  `json:"timestamp"`
}
