package common

import (
	"math"
)

func SafeRoundedFloat(val float64) float64 {
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return 0.0
	}

	return math.Round(val*1000) / 1000
}
