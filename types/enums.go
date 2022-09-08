package types

import (
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

func (t PeriodType) Window() int {
	return int(t.Average() / t.Rollup())
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
