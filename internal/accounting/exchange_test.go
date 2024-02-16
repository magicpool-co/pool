package accounting

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/magicpool-co/pool/pkg/common"
)

var (
	priceIndex = map[string]map[string]float64{
		"CFX": map[string]float64{
			"BTC":  0.00000243,
			"ETH":  0.00003344,
			"USDC": 0.04897981,
		},
		"ERG": map[string]float64{
			"BTC":  0.00020660,
			"ETH":  0.00283834,
			"USDC": 4.16,
		},
		"ETC": map[string]float64{
			"BTC":  0.00171175,
			"ETH":  0.02353354,
			"USDC": 34.46,
		},
		"FIRO": map[string]float64{
			"BTC":  0.00014882,
			"ETH":  0.00204601,
			"USDC": 3.00,
		},
		"FLUX": map[string]float64{
			"BTC":  0.00005442,
			"ETH":  0.00074827,
			"USDC": 1.10,
		},
		"RVN": map[string]float64{
			"BTC":  0.00000241,
			"ETH":  0.00003305,
			"USDC": 0.04865315,
		},
	}
)

func TestReverseMap(t *testing.T) {
	tests := []struct {
		input  map[string]map[string]*big.Int
		prices map[string]map[string]float64
		output map[string]map[string]*big.Int
	}{
		{
			input: map[string]map[string]*big.Int{
				"CFX": map[string]*big.Int{
					"ETH":  common.MustParseBigInt("2000000000000000000000"),
					"BTC":  common.MustParseBigInt("2000000000000000000000"),
					"USDC": common.MustParseBigInt("2000000000000000000000"),
				},
				"ERG": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(10_000_000_000),
					"BTC":  new(big.Int).SetUint64(10_000_000_000),
					"USDC": new(big.Int).SetUint64(10_000_000_000),
				},
				"ETC": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(3_000_000_000_000_000_000),
					"BTC":  new(big.Int).SetUint64(3_000_000_000_000_000_000),
					"USDC": new(big.Int).SetUint64(3_000_000_000_000_000_000),
				},
				"FIRO": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(10_000_000_000),
					"BTC":  new(big.Int).SetUint64(10_000_000_000),
					"USDC": new(big.Int).SetUint64(10_000_000_000),
				},
				"FLUX": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(30_000_000_000),
					"BTC":  new(big.Int).SetUint64(30_000_000_000),
					"USDC": new(big.Int).SetUint64(30_000_000_000),
				},
				"RVN": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(20_000_000_000),
					"BTC":  new(big.Int).SetUint64(20_000_000_000),
					"USDC": new(big.Int).SetUint64(20_000_000_000),
				},
			},
			prices: priceIndex,
			output: map[string]map[string]*big.Int{
				"BTC": map[string]*big.Int{
					"CFX":  new(big.Int).SetUint64(486000),
					"ERG":  new(big.Int).SetUint64(206600),
					"ETC":  new(big.Int).SetUint64(513525),
					"FIRO": new(big.Int).SetUint64(1488200),
					"FLUX": new(big.Int).SetUint64(1632600),
					"RVN":  new(big.Int).SetUint64(48200),
				},
				"ETH": map[string]*big.Int{
					"CFX":  new(big.Int).SetUint64(66880000000000000),
					"ERG":  new(big.Int).SetUint64(28383400000000000),
					"ETC":  new(big.Int).SetUint64(70600620000000000),
					"FIRO": new(big.Int).SetUint64(204601000000000000),
					"FLUX": new(big.Int).SetUint64(224481000000000000),
					"RVN":  new(big.Int).SetUint64(6610000000000000),
				},
				"USDC": map[string]*big.Int{
					"CFX":  new(big.Int).SetUint64(97958000),
					"ERG":  new(big.Int).SetUint64(41600000),
					"ETC":  new(big.Int).SetUint64(103380000),
					"FIRO": new(big.Int).SetUint64(300000000),
					"FLUX": new(big.Int).SetUint64(330000000),
					"RVN":  new(big.Int).SetUint64(9730600),
				},
			},
		},
		{
			input: map[string]map[string]*big.Int{
				"KAS": map[string]*big.Int{
					"BTC": new(big.Int).SetUint64(332442649077),
					"ETH": new(big.Int).SetUint64(1711990761540),
				},
				"NEXA": map[string]*big.Int{
					"ETH": new(big.Int).SetUint64(80633842496),
				},
			},
			prices: map[string]map[string]float64{
				"KAS": map[string]float64{
					"BTC": 7.968882455859694e-07,
					"ETH": 1.2952496931210139e-05,
				},
				"NEXA": map[string]float64{
					"ETH": 2.9193848389144937e-09,
				},
			},
			output: map[string]map[string]*big.Int{
				"BTC": map[string]*big.Int{
					"KAS": new(big.Int).SetUint64(262629),
				},
				"ETH": map[string]*big.Int{
					"KAS":  new(big.Int).SetUint64(221745550830352120),
					"NEXA": new(big.Int).SetUint64(2354008271059724800),
				},
			},
		},
	}

	for i, tt := range tests {
		output, err := reverseMap(tt.input, tt.prices)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !reflect.DeepEqual(output, tt.output) {
			t.Errorf("failed on %d: output mismatch: have %v, want %v", i, output, tt.output)
		}
	}
}

func TestSumMap(t *testing.T) {
	tests := []struct {
		input  map[string]map[string]*big.Int
		output map[string]map[string]*big.Int
	}{}

	for i, tt := range tests {
		output := sumMap(tt.input)
		if !reflect.DeepEqual(output, tt.output) {
			t.Errorf("failed on %d: output mismatch: have %v, want %v", i, output, tt.output)
		}
	}
}

func TestCalculateExchangePaths(t *testing.T) {
	tests := []struct {
		inputPaths       map[string]map[string]*big.Int
		inputThresholds  map[string]*big.Int
		outputThresholds map[string]*big.Int
		prices           map[string]map[string]float64
		finalPaths       map[string]map[string]*big.Int
	}{
		{
			inputPaths: map[string]map[string]*big.Int{
				"CFX": map[string]*big.Int{
					"ETH": common.MustParseBigInt("7433150323012392030309"),
					"BTC": common.MustParseBigInt("4225155251235918477239"),
				},
				"ETC": map[string]*big.Int{
					"ETH":  common.MustParseBigInt("82315931231311938231"),
					"BTC":  common.MustParseBigInt("31412030881418410073"),
					"USDC": common.MustParseBigInt("39310403813440000003"),
				},
				"FLUX": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(18_301_031_941),
					"USDC": new(big.Int).SetUint64(14_299_031_132),
				},
			},
			inputThresholds: DefaultInputThresholds,
			outputThresholds: map[string]*big.Int{
				"BTC":  new(big.Int).SetUint64(5_000_000),
				"ETH":  new(big.Int).SetUint64(500_000_000_000_000_000),
				"USDC": new(big.Int).SetUint64(2_000_000_000),
			},
			prices: priceIndex,
			finalPaths: map[string]map[string]*big.Int{
				"CFX": map[string]*big.Int{
					"ETH": common.MustParseBigInt("7433150323012392030309"),
					"BTC": common.MustParseBigInt("4225155251235918477239"),
				},
				"ETC": map[string]*big.Int{
					"ETH": common.MustParseBigInt("82315931231311938231"),
					"BTC": common.MustParseBigInt("31412030881418410073"),
				},
			},
		},
		{
			inputPaths: map[string]map[string]*big.Int{
				"CFX": map[string]*big.Int{
					"ETH":  common.MustParseBigInt("2000000000000000000000"),
					"BTC":  common.MustParseBigInt("2000000000000000000000"),
					"USDC": common.MustParseBigInt("2000000000000000000000"),
				},
				"ERG": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(39_049_076_512_513),
					"BTC":  new(big.Int).SetUint64(241_000_423_000_312),
					"USDC": new(big.Int).SetUint64(132_042_004_041_420_314),
				},
				"ETC": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(3_000_000_000_000_000_000),
					"BTC":  new(big.Int).SetUint64(3_000_000_000_000_000_000),
					"USDC": new(big.Int).SetUint64(3_000_000_000_000_000_000),
				},
				"FIRO": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(10_000_000_000),
					"BTC":  new(big.Int).SetUint64(10_000_000_000),
					"USDC": new(big.Int).SetUint64(10_000_000_000),
				},
				"FLUX": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(30_000_000_000),
					"BTC":  new(big.Int).SetUint64(30_000_000_000),
					"USDC": new(big.Int).SetUint64(30_000_000_000),
				},
				"RVN": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(20_000_000_000),
					"BTC":  new(big.Int).SetUint64(20_000_000_000),
					"USDC": new(big.Int).SetUint64(20_000_000_000),
				},
			},
			inputThresholds: DefaultInputThresholds,
			outputThresholds: map[string]*big.Int{
				"BTC":  new(big.Int).SetUint64(50_000_000),
				"ETH":  new(big.Int).SetUint64(5_000_000_000_000_000_000),
				"USDC": new(big.Int).SetUint64(20_000_000_000),
			},
			prices: priceIndex,
			finalPaths: map[string]map[string]*big.Int{
				"CFX": map[string]*big.Int{
					"ETH":  common.MustParseBigInt("2000000000000000000000"),
					"BTC":  common.MustParseBigInt("2000000000000000000000"),
					"USDC": common.MustParseBigInt("2000000000000000000000"),
				},
				"ERG": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(39_049_076_512_513),
					"BTC":  new(big.Int).SetUint64(241_000_423_000_312),
					"USDC": new(big.Int).SetUint64(132_042_004_041_420_314),
				},
				"ETC": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(3_000_000_000_000_000_000),
					"BTC":  new(big.Int).SetUint64(3_000_000_000_000_000_000),
					"USDC": new(big.Int).SetUint64(3_000_000_000_000_000_000),
				},
				"FIRO": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(10_000_000_000),
					"BTC":  new(big.Int).SetUint64(10_000_000_000),
					"USDC": new(big.Int).SetUint64(10_000_000_000),
				},
				"FLUX": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(30_000_000_000),
					"BTC":  new(big.Int).SetUint64(30_000_000_000),
					"USDC": new(big.Int).SetUint64(30_000_000_000),
				},
				"RVN": map[string]*big.Int{
					"ETH":  new(big.Int).SetUint64(20_000_000_000),
					"BTC":  new(big.Int).SetUint64(20_000_000_000),
					"USDC": new(big.Int).SetUint64(20_000_000_000),
				},
			},
		},
		{
			inputPaths: map[string]map[string]*big.Int{
				"KAS": map[string]*big.Int{
					"BTC": new(big.Int).SetUint64(332442649077),
					"ETH": new(big.Int).SetUint64(1711990761540),
				},
				"NEXA": map[string]*big.Int{
					"ETH": new(big.Int).SetUint64(80633842496),
				},
			},
			inputThresholds: map[string]*big.Int{
				"CFX":  common.MustParseBigInt("2000000000000000000000000"),
				"ERG":  new(big.Int).SetUint64(100_000_000_000),
				"ETC":  common.MustParseBigInt("25000000000000000000"),
				"KAS":  new(big.Int).SetUint64(100_000_000_000),
				"FIRO": new(big.Int).SetUint64(10_000_000_000),
				"FLUX": new(big.Int).SetUint64(10_000_000_000),
				"NEXA": new(big.Int).SetUint64(100_000_000),
				"RVN":  new(big.Int).SetUint64(500_000_000_000),
			},
			outputThresholds: map[string]*big.Int{
				"BTC":  new(big.Int).SetUint64(2_500_000),
				"ETH":  new(big.Int).SetUint64(250_000_000_000_000_000),
				"USDC": new(big.Int).SetUint64(20_000_000_000),
			},
			prices: map[string]map[string]float64{
				"KAS": map[string]float64{
					"BTC": 7.968882455859694e-07,
					"ETH": 1.2952496931210139e-05,
				},
				"NEXA": map[string]float64{
					"ETH": 2.9193848389144937e-09,
				},
			},
			finalPaths: map[string]map[string]*big.Int{
				"KAS": map[string]*big.Int{
					"ETH": common.MustParseBigInt("1711990761540"),
				},
				"NEXA": map[string]*big.Int{
					"ETH": common.MustParseBigInt("80633842496"),
				},
			},
		},
	}

	for i, tt := range tests {
		finalPaths, err := CalculateExchangePaths(tt.inputPaths,
			tt.inputThresholds, tt.outputThresholds, tt.prices)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !reflect.DeepEqual(finalPaths, tt.finalPaths) {
			t.Errorf("failed on %d: final paths mismatch: have %v, want %v",
				i, finalPaths, tt.finalPaths)
		}
	}
}

func TestCalculateProportionalValues(t *testing.T) {
	tests := []struct {
		value              *big.Int
		fee                *big.Int
		proportions        map[string]*big.Int
		proportionalValues map[string]*big.Int
		proportionalFees   map[string]*big.Int
	}{
		{
			value: new(big.Int).SetUint64(3191468400000000000),
			fee:   new(big.Int).SetUint64(8400456920411368),
			proportions: map[string]*big.Int{
				"ETC":  new(big.Int).SetUint64(2997019000000000000),
				"FLUX": new(big.Int).SetUint64(199449400000000000),
			},
			proportionalValues: map[string]*big.Int{
				"ETC":  new(big.Int).SetUint64(2992330984000842931),
				"FLUX": new(big.Int).SetUint64(199137415999157069),
			},
			proportionalFees: map[string]*big.Int{
				"ETC":  new(big.Int).SetUint64(7876295288623645),
				"FLUX": new(big.Int).SetUint64(524161631787723),
			},
		},
		{
			value: new(big.Int).SetUint64(7757400),
			fee:   new(big.Int).SetUint64(57983),
			proportions: map[string]*big.Int{
				"ETC":  new(big.Int).SetUint64(6132400),
				"FLUX": new(big.Int).SetUint64(1675000),
			},
			proportionalValues: map[string]*big.Int{
				"ETC":  new(big.Int).SetUint64(6093128),
				"FLUX": new(big.Int).SetUint64(1664272),
			},
			proportionalFees: map[string]*big.Int{
				"ETC":  new(big.Int).SetUint64(45544),
				"FLUX": new(big.Int).SetUint64(12439),
			},
		},
		{
			value: new(big.Int).SetUint64(39_049_076_512_513),
			fee:   new(big.Int).SetUint64(40_139_932_481),
			proportions: map[string]*big.Int{
				"a": new(big.Int).SetUint64(100),
			},
			proportionalValues: map[string]*big.Int{
				"a": new(big.Int).SetUint64(39_049_076_512_513),
			},
			proportionalFees: map[string]*big.Int{
				"a": new(big.Int).SetUint64(40_139_932_481),
			},
		},
		{
			value: new(big.Int).SetUint64(39_049_076_512_513),
			fee:   new(big.Int).SetUint64(40_139_932_481),
			proportions: map[string]*big.Int{
				"a": new(big.Int).SetUint64(50),
				"b": new(big.Int).SetUint64(50),
			},
			proportionalValues: map[string]*big.Int{
				"a": new(big.Int).SetUint64(19_524_538_256_257),
				"b": new(big.Int).SetUint64(19_524_538_256_256),
			},
			proportionalFees: map[string]*big.Int{
				"a": new(big.Int).SetUint64(20_069_966_241),
				"b": new(big.Int).SetUint64(20_069_966_240),
			},
		},
		{
			value: new(big.Int).SetUint64(39_049_076_512_513),
			fee:   new(big.Int).SetUint64(40_139_932_481),
			proportions: map[string]*big.Int{
				"a": new(big.Int).SetUint64(2_135_123),
				"b": new(big.Int).SetUint64(51_235_123_125_123),
				"c": new(big.Int).SetUint64(59_840_203_041),
				"d": new(big.Int).SetUint64(32),
				"e": new(big.Int).SetUint64(258_881_824_858_293),
			},
			proportionalValues: map[string]*big.Int{
				"a": new(big.Int).SetUint64(268_796),
				"b": new(big.Int).SetUint64(6_450_141_678_771),
				"c": new(big.Int).SetUint64(7_533_460_722),
				"d": new(big.Int).SetUint64(4),
				"e": new(big.Int).SetUint64(32_591_401_104_220),
			},
			proportionalFees: map[string]*big.Int{
				"a": new(big.Int).SetUint64(276),
				"b": new(big.Int).SetUint64(6_630_329_692),
				"c": new(big.Int).SetUint64(7_743_911),
				"d": new(big.Int).SetUint64(0),
				"e": new(big.Int).SetUint64(33_501_858_602),
			},
		},
		{
			value: new(big.Int).SetUint64(40_139_932_481),
			fee:   new(big.Int).SetUint64(39_049_076_512_513),
			proportions: map[string]*big.Int{
				"a": new(big.Int).SetUint64(2_135_123),
				"b": new(big.Int).SetUint64(51_235_123_125_123),
				"c": new(big.Int).SetUint64(59_840_203_041),
				"d": new(big.Int).SetUint64(32),
				"e": new(big.Int).SetUint64(258_881_824_858_293),
			},
			proportionalValues: map[string]*big.Int{
				"a": new(big.Int).SetUint64(276),
				"b": new(big.Int).SetUint64(6_630_329_692),
				"c": new(big.Int).SetUint64(7_743_911),
				"d": new(big.Int).SetUint64(0),
				"e": new(big.Int).SetUint64(33_501_858_602),
			},
			proportionalFees: map[string]*big.Int{
				"a": new(big.Int).SetUint64(268_796),
				"b": new(big.Int).SetUint64(6_450_141_678_771),
				"c": new(big.Int).SetUint64(7_533_460_722),
				"d": new(big.Int).SetUint64(4),
				"e": new(big.Int).SetUint64(32_591_401_104_220),
			},
		},
	}

	for i, tt := range tests {
		values, fees, err := CalculateProportionalValues(tt.value, tt.fee, tt.proportions)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !common.DeepEqualMapBigInt1D(values, tt.proportionalValues) {
			t.Errorf("failed on %d: proportional values mismatch: have %v, want %v",
				i, values, tt.proportionalValues)
		} else if !common.DeepEqualMapBigInt1D(fees, tt.proportionalFees) {
			t.Errorf("failed on %d: proportional fees mismatch: have %v, want %v",
				i, fees, tt.proportionalFees)
		}
	}
}
