package charter

import (
	"math"
)

var (
	chains = []string{"CFX", "CTXC", "ERGO", "ETC", "FIRO", "FLUX", "RVN"}
)

func processFloat(val float64) float64 {
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return 0.0
	}

	return math.Round(val*1000) / 1000
}
