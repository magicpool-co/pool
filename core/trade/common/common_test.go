// Copyright © 2022 Igor Kroitor

package common

import (
	"testing"
)

func TestDecimalToPrecision(t *testing.T) {
	tests := []struct {
		input     string
		rounding  RoundingMode
		precision string
		digits    DigitsMode
		padding   PaddingMode
		output    string
	}{
		{"12.3456000", Truncate, "100", DecimalPlaces, NoPadding, "12.3456"},
		{"12.3456", Truncate, "100", DecimalPlaces, NoPadding, "12.3456"},
		{"12.3456", Truncate, "4", DecimalPlaces, NoPadding, "12.3456"},
		{"12.3456", Truncate, "3", DecimalPlaces, NoPadding, "12.345"},
		{"12.3456", Truncate, "2", DecimalPlaces, NoPadding, "12.34"},
		{"12.3456", Truncate, "1", DecimalPlaces, NoPadding, "12.3"},
		{"12.3456", Truncate, "0", DecimalPlaces, NoPadding, "12"},
		{"0.0000001", Truncate, "8", DecimalPlaces, NoPadding, "0.0000001"},
		{"0.00000001", Truncate, "8", DecimalPlaces, NoPadding, "0.00000001"},
		{"0.000000000", Truncate, "9", DecimalPlaces, PadWithZero, "0.000000000"},
		{"0.000000001", Truncate, "9", DecimalPlaces, PadWithZero, "0.000000001"},
		{"12.3456", Truncate, "-1", DecimalPlaces, NoPadding, "10"},
		{"123.456", Truncate, "-1", DecimalPlaces, NoPadding, "120"},
		{"123.456", Truncate, "-2", DecimalPlaces, NoPadding, "100"},
		{"9.99999", Truncate, "-1", DecimalPlaces, NoPadding, "0"},
		{"99.9999", Truncate, "-1", DecimalPlaces, NoPadding, "90"},
		{"99.9999", Truncate, "-2", DecimalPlaces, NoPadding, "0"},
		{"0", Truncate, "0", DecimalPlaces, NoPadding, "0"},
		{"-0.9", Truncate, "0", DecimalPlaces, NoPadding, "0"},
		{"0.000123456700", Truncate, "100", SignificantDigits, NoPadding, "0.0001234567"},
		{"0.0001234567", Truncate, "100", SignificantDigits, NoPadding, "0.0001234567"},
		{"0.0001234567", Truncate, "7", SignificantDigits, NoPadding, "0.0001234567"},
		{"0.000123456", Truncate, "6", SignificantDigits, NoPadding, "0.000123456"},
		{"0.000123456", Truncate, "5", SignificantDigits, NoPadding, "0.00012345"},
		{"0.000123456", Truncate, "2", SignificantDigits, NoPadding, "0.00012"},
		{"0.000123456", Truncate, "1", SignificantDigits, NoPadding, "0.0001"},
		{"123.0000987654", Truncate, "10", SignificantDigits, PadWithZero, "123.0000987"},
		{"123.0000987654", Truncate, "8", SignificantDigits, NoPadding, "123.00009"},
		{"123.0000987654", Truncate, "7", SignificantDigits, PadWithZero, "123.0000"},
		{"123.0000987654", Truncate, "6", SignificantDigits, NoPadding, "123"},
		{"123.0000987654", Truncate, "5", SignificantDigits, PadWithZero, "123.00"},
		{"123.0000987654", Truncate, "4", SignificantDigits, NoPadding, "123"},
		{"123.0000987654", Truncate, "4", SignificantDigits, PadWithZero, "123.0"},
		{"123.0000987654", Truncate, "3", SignificantDigits, PadWithZero, "123"},
		{"123.0000987654", Truncate, "2", SignificantDigits, NoPadding, "120"},
		{"123.0000987654", Truncate, "1", SignificantDigits, NoPadding, "100"},
		{"123.0000987654", Truncate, "1", SignificantDigits, PadWithZero, "100"},
		{"1234", Truncate, "5", SignificantDigits, NoPadding, "1234"},
		{"1234", Truncate, "5", SignificantDigits, PadWithZero, "1234.0"},
		{"1234", Truncate, "4", SignificantDigits, NoPadding, "1234"},
		{"1234", Truncate, "4", SignificantDigits, PadWithZero, "1234"},
		{"1234.69", Truncate, "0", SignificantDigits, NoPadding, "0"},
		{"1234.69", Truncate, "0", SignificantDigits, PadWithZero, "0"},
		{"12.3456000", Round, "100", DecimalPlaces, NoPadding, "12.3456"},
		{"12.3456", Round, "100", DecimalPlaces, NoPadding, "12.3456"},
		{"12.3456", Round, "4", DecimalPlaces, NoPadding, "12.3456"},
		{"12.3456", Round, "3", DecimalPlaces, NoPadding, "12.346"},
		{"12.3456", Round, "2", DecimalPlaces, NoPadding, "12.35"},
		{"12.3456", Round, "1", DecimalPlaces, NoPadding, "12.3"},
		{"12.3456", Round, "0", DecimalPlaces, NoPadding, "12"},
		{"10000", Round, "6", DecimalPlaces, NoPadding, "10000"},
		{"0.00003186", Round, "8", DecimalPlaces, NoPadding, "0.00003186"},
		{"12.3456", Round, "-1", DecimalPlaces, NoPadding, "10"},
		{"123.456", Round, "-1", DecimalPlaces, NoPadding, "120"},
		{"123.456", Round, "-2", DecimalPlaces, NoPadding, "100"},
		{"9.99999", Round, "-1", DecimalPlaces, NoPadding, "10"},
		{"99.9999", Round, "-1", DecimalPlaces, NoPadding, "100"},
		{"99.9999", Round, "-2", DecimalPlaces, NoPadding, "100"},
		{"9.999", Round, "3", DecimalPlaces, NoPadding, "9.999"},
		{"9.999", Round, "2", DecimalPlaces, NoPadding, "10"},
		{"9.999", Round, "2", DecimalPlaces, PadWithZero, "10.00"},
		{"99.999", Round, "2", DecimalPlaces, PadWithZero, "100.00"},
		{"-99.999", Round, "2", DecimalPlaces, PadWithZero, "-100.00"},
		{"0.000123456700", Round, "100", SignificantDigits, NoPadding, "0.0001234567"},
		{"0.0001234567", Round, "100", SignificantDigits, NoPadding, "0.0001234567"},
		{"0.0001234567", Round, "7", SignificantDigits, NoPadding, "0.0001234567"},
		{"0.000123456", Round, "6", SignificantDigits, NoPadding, "0.000123456"},
		{"0.000123456", Round, "5", SignificantDigits, NoPadding, "0.00012346"},
		{"0.000123456", Round, "4", SignificantDigits, NoPadding, "0.0001235"},
		{"0.00012", Round, "2", SignificantDigits, NoPadding, "0.00012"},
		{"0.0001", Round, "1", SignificantDigits, NoPadding, "0.0001"},
		{"123.0000987654", Round, "7", SignificantDigits, NoPadding, "123.0001"},
		{"123.0000987654", Round, "6", SignificantDigits, NoPadding, "123"},
		{"0.00098765", Round, "2", SignificantDigits, NoPadding, "0.00099"},
		{"0.00098765", Round, "2", SignificantDigits, PadWithZero, "0.00099"},
		{"0.00098765", Round, "1", SignificantDigits, NoPadding, "0.001"},
		{"0.00098765", Round, "10", SignificantDigits, PadWithZero, "0.0009876500000"},
		{"0.098765", Round, "1", SignificantDigits, PadWithZero, "0.1"},
		{"0", Round, "0", SignificantDigits, NoPadding, "0"},
		{"-0.123", Round, "0", SignificantDigits, NoPadding, "0"},
		{"0.00000044", Round, "5", SignificantDigits, NoPadding, "0.00000044"},
		{"0.000123456700", Round, "0.00012", TickSize, NoPadding, "0.00012"},
		{"0.0001234567", Round, "0.00013", TickSize, NoPadding, "0.00013"},
		{"0.0001234567", Truncate, "0.00013", TickSize, NoPadding, "0"},
		{"101.000123456700", Round, "100", TickSize, NoPadding, "100"},
		{"0.000123456700", Round, "100", TickSize, NoPadding, "0"},
		{"165", Truncate, "110", TickSize, NoPadding, "110"},
		{"3210", Truncate, "1110", TickSize, NoPadding, "2220"},
		{"165", Round, "110", TickSize, NoPadding, "220"},
		{"0.000123456789", Round, "0.00000012", TickSize, NoPadding, "0.00012348"},
		{"0.000123456789", Truncate, "0.00000012", TickSize, NoPadding, "0.00012336"},
		{"0.000273398", Round, "1e-7", TickSize, NoPadding, "0.0002734"},
		{"0.00005714", Truncate, "0.00000001", TickSize, NoPadding, "0.00005714"},
		{"0.0000571495257361", Truncate, "0.00000001", TickSize, NoPadding, "0.00005714"},
		{"0.01", Round, "0.0001", TickSize, PadWithZero, "0.0100"},
		{"0.01", Truncate, "0.0001", TickSize, PadWithZero, "0.0100"},
		{"-0.000123456789", Round, "0.00000012", TickSize, NoPadding, "-0.00012348"},
		{"-0.000123456789", Truncate, "0.00000012", TickSize, NoPadding, "-0.00012336"},
		{"-165", Truncate, "110", TickSize, NoPadding, "-110"},
		{"-165", Round, "110", TickSize, NoPadding, "-220"},
		{"-1650", Truncate, "1100", TickSize, NoPadding, "-1100"},
		{"-1650", Round, "1100", TickSize, NoPadding, "-2200"},
		{"0.0006", Truncate, "0.0001", TickSize, NoPadding, "0.0006"},
		{"-0.0006", Truncate, "0.0001", TickSize, NoPadding, "-0.0006"},
		{"0.6", Truncate, "0.2", TickSize, NoPadding, "0.6"},
		{"-0.6", Truncate, "0.2", TickSize, NoPadding, "-0.6"},
		{"1.2", Round, "0.4", TickSize, NoPadding, "1.2"},
		{"-1.2", Round, "0.4", TickSize, NoPadding, "-1.2"},
		{"1.2", Round, "0.02", TickSize, NoPadding, "1.2"},
		{"-1.2", Round, "0.02", TickSize, NoPadding, "-1.2"},
		{"44", Round, "4.4", TickSize, NoPadding, "44"},
		{"-44", Round, "4.4", TickSize, NoPadding, "-44"},
		{"44.00000001", Round, "4.4", TickSize, NoPadding, "44"},
		{"-44.00000001", Round, "4.4", TickSize, NoPadding, "-44"},
		{"20", Truncate, "0.00000001", TickSize, NoPadding, "20"},
		{"-0.123456", Truncate, "5", DecimalPlaces, NoPadding, "-0.12345"},
		{"-0.123456", Round, "5", DecimalPlaces, NoPadding, "-0.12346"},
		{"123", Truncate, "0", DecimalPlaces, NoPadding, "123"},
		{"123", Truncate, "5", DecimalPlaces, NoPadding, "123"},
		{"123", Truncate, "5", DecimalPlaces, PadWithZero, "123.00000"},
		{"123.", Truncate, "0", DecimalPlaces, NoPadding, "123"},
		{"123.", Truncate, "5", DecimalPlaces, PadWithZero, "123.00000"},
		{"0.", Truncate, "0", DecimalPlaces, NoPadding, "0"},
		{"0.", Truncate, "5", DecimalPlaces, PadWithZero, "0.00000"},
		{"1.44", Round, "1", DecimalPlaces, NoPadding, "1.4"},
		{"1.45", Round, "1", DecimalPlaces, NoPadding, "1.5"},
		{"1.45", Round, "0", DecimalPlaces, NoPadding, "1"},
		{"5", Round, "-1", DecimalPlaces, NoPadding, "10"},
		{"4.999", Round, "-1", DecimalPlaces, NoPadding, "0"},
		{"0.0431531423", Round, "-1", DecimalPlaces, NoPadding, "0"},
		{"-69.3", Round, "-1", DecimalPlaces, NoPadding, "-70"},
		{"5001", Round, "-4", DecimalPlaces, NoPadding, "10000"},
		{"4999.999", Round, "-4", DecimalPlaces, NoPadding, "0"},
		{"69.3", Truncate, "-2", DecimalPlaces, NoPadding, "0"},
		{"-69.3", Truncate, "-2", DecimalPlaces, NoPadding, "0"},
		{"69.3", Truncate, "-1", SignificantDigits, NoPadding, "60"},
		{"-69.3", Truncate, "-1", SignificantDigits, NoPadding, "-60"},
		{"69.3", Truncate, "-2", SignificantDigits, NoPadding, "0"},
		{"1602000000000000000000", Truncate, "3", SignificantDigits, NoPadding, "1600000000000000000000"},
	}

	for i, tt := range tests {
		output, err := DecimalToPrecision(tt.input, tt.precision, tt.rounding, tt.digits, tt.padding)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if output != tt.output {
			t.Errorf("failed on %d: have %s, want %s", i, output, tt.output)
		}
	}
}
