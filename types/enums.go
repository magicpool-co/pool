package types

import (
	"fmt"
	"strings"
	"time"
)

type ShareStatus int

const (
	AcceptedShare ShareStatus = iota
	RejectedShare
	InvalidShare
)

type PeriodType int

const (
	Period15m PeriodType = iota
	Period1h
	Period4h
	Period1d
)

func ParsePeriodType(raw string) (PeriodType, error) {
	switch strings.ToLower(raw) {
	case "15m":
		return Period15m, nil
	case "1h":
		return Period1h, nil
	case "4h":
		return Period4h, nil
	case "1d":
		return Period1d, nil
	default:
		return 0, fmt.Errorf("invalid period type")
	}
}

func (t PeriodType) AverageWindow() int {
	return int(t.Average() / t.Rollup())
}

func (t PeriodType) RetentionWindow() int {
	return int(t.Retention() / t.Rollup())
}

func (t PeriodType) Rollup() time.Duration {
	switch t {
	case 0:
		return time.Minute * 15
	case 1:
		return time.Hour
	case 2:
		return time.Hour * 4
	case 3:
		return time.Hour * 24
	default:
		return time.Minute
	}
}

func (t PeriodType) Average() time.Duration {
	switch t {
	case 0:
		return time.Hour * 4
	case 1:
		return time.Hour * 8
	case 2:
		return time.Hour * 24
	case 3:
		return time.Hour * 24 * 30
	default:
		return time.Hour
	}
}

func (t PeriodType) Retention() time.Duration {
	switch t {
	case 0:
		return time.Hour * 24
	case 1:
		return time.Hour * 24 * 7
	case 2:
		return time.Hour * 24 * 30
	case 3:
		return time.Hour * 24 * 365
	default:
		return time.Hour
	}
}

func (t PeriodType) GenerateRange(endTime time.Time) map[time.Time]bool {
	index := make(map[time.Time]bool)
	for i := t.RetentionWindow(); i >= 0; i-- {
		entry := endTime.Add(-t.Rollup() * time.Duration(i))
		index[entry] = false
	}

	return index
}
