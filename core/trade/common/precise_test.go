// Copyright © 2022 Igor Kroitor

package common

import (
	"fmt"
	"testing"
)

func TestPreciseStringOperators(t *testing.T) {
	tests := []struct {
		a     string
		b     string
		op    string
		value string
	}{
		{"0.00000002", "69696900000", "+", "69696900000.00000002"},
		{"69696900000", "0.00000002", "+", "69696900000.00000002"},
		{"0.00000002", "-1.123e-6", "+", "-0.000001103"},
		{"-1.123e-6", "0.00000002", "+", "-0.000001103"},
		{"0", "-1.123e-6", "+", "-0.000001123"},
		{"0", "0.00000002", "+", "0.00000002"},
		{"0", "69696900000", "+", "69696900000"},
		{"-1.123e-6", "0", "+", "-0.000001123"},
		{"0.00000002", "0", "+", "0.00000002"},
		{"69696900000", "0", "+", "69696900000"},

		{"0.00000002", "69696900000", "-", "-69696899999.99999998"},
		{"69696900000", "0.00000002", "-", "69696899999.99999998"},
		{"0.00000002", "-1.123e-6", "-", "0.000001143"},
		{"-1.123e-6", "0.00000002", "-", "-0.000001143"},

		{"0.00000002", "69696900000", "*", "1393.938"},
		{"69696900000", "0.00000002", "*", "1393.938"},
		{"0.00000002", "-1.123e-6", "*", "-0.00000000000002246"},
		{"-1.123e-6", "0.00000002", "*", "-0.00000000000002246"},
		{"0", "-1.123e-6", "*", "0"},
		{"0", "0.00000002", "*", "0"},
		{"0", "69696900000", "*", "0"},
		{"-1.123e-6", "0", "*", "0"},
		{"0.00000002", "0", "*", "0"},
		{"69696900000", "0", "*", "0"},
		{"0.00000002", "1e8", "*", "2"},
		{"1e8", "0.00000002", "*", "2"},
		{"69696900000", "1e8", "*", "6969690000000000000"},
		{"1e8", "69696900000", "*", "6969690000000000000"},

		{"0.00000002", "69696900000", "/", "0"},
		{"69696900000", "0.00000002", "/", "3484845000000000000"},
		{"0.00000002", "-1.123e-6", "/", "-0.017809439002671415"},
		{"-1.123e-6", "0.00000002", "/", "-56.15"},
		{"69696900000", "1e8", "/", "696.969"},
		{"1e8", "69696900000", "/", "0.001434784043479695"},
	}

	for i, tt := range tests {
		var value string
		var err error
		switch tt.op {
		case "+":
			value, err = PreciseStringAdd(tt.a, tt.b)
		case "-":
			value, err = PreciseStringSub(tt.a, tt.b)
		case "*":
			value, err = PreciseStringMul(tt.a, tt.b)
		case "/":
			value, err = PreciseStringDiv(tt.a, tt.b)
		default:
			err = fmt.Errorf("unknown operator")
		}

		if err != nil {
			t.Errorf("failed on %d: %s: %v", i, tt.op, err)
		} else if value != tt.value {
			t.Errorf("failed on %d: %s:, have %s, want %s", i, tt.op, value, tt.value)
		}
	}
}

func TestPreciseStringDivWithPrecision(t *testing.T) {
	tests := []struct {
		a         string
		b         string
		precision int
		value     string
	}{
		{"0.00000002", "69696900000", 1, "0"},
		{"0.00000002", "69696900000", 19, "0.0000000000000000002"},
		{"0.00000002", "69696900000", 20, "0.00000000000000000028"},
		{"0.00000002", "69696900000", 21, "0.000000000000000000286"},
		{"0.00000002", "69696900000", 22, "0.0000000000000000002869"},
		{"69696900000", "1e8", -1, "690"},
		{"69696900000", "1e8", 0, "696"},
		{"69696900000", "1e8", 1, "696.9"},
		{"69696900000", "1e8", 2, "696.96"},
		{"72.76171481", "0.012824", 6, "5673.870462"},
	}

	for i, tt := range tests {
		value, err := PreciseStringDivWithPrecision(tt.a, tt.b, tt.precision)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if value != tt.value {
			t.Errorf("failed on %d: have %s, want %s", i, value, tt.value)
		}
	}
}

func TestPreciseAdjustedDigits(t *testing.T) {
	tests := []struct {
		input  string
		digits int
	}{
		{"0", 0},
		{"-0", 0},
		{"69", 1},
		{"-69", 1},
		{"1234e9999", 10002},
		{"-1234e9999", 10002},
		{"35018421976.794013996282", 10},
		{"-35018421976.794013996282", 10},
		{"7706.492775", 3},
		{"987654", 5},
		{"98765.4", 4},
		{"9876.54", 3},
		{"987.654", 2},
		{"98.7654", 1},
		{"9.87654", 0},
		{"0.987654", -1},
		{"0.0987654", -2},
		{"0.00987654", -3},
		{"0.000987654", -4},
		{"0.0000987654", -5},
		{"0.000003412", -6},
	}

	for i, tt := range tests {
		precise, err := NewPrecise(tt.input)
		if err != nil {
			t.Errorf("failed on %d: NewPrecise: %v", i, err)
			continue
		}

		digits := precise.AdjustedDigits()
		if digits != tt.digits {
			t.Errorf("failed on %d: have %d, want %d", i, digits, tt.digits)
		}
	}
}

func TestPreciseAbs(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"0", "0"},
		{"-0", "0"},
		{"-500.1", "500.1"},
		{"213", "213"},
	}

	for i, tt := range tests {
		precise, err := NewPrecise(tt.input)
		if err != nil {
			t.Errorf("failed on %d: NewPrecise: %v", i, err)
			continue
		}

		output := precise.Abs().String()
		if output != tt.output {
			t.Errorf("failed on %d: have %s, want %s", i, output, tt.output)
		}
	}
}

func TestPreciseComparators(t *testing.T) {
	tests := []struct {
		a      string
		b      string
		op     string
		output bool
	}{

		{"1.0000", "2", "gt", false},
		{"2", "1.2345", "gt", true},
		{"3.1415", "-2", "gt", true},
		{"-3.1415", "-2", "gt", false},
		{"3.1415", "3.1415", "gt", false},
		{"3.14150000000000000000001", "3.1415", "gt", true},

		{"1.0000", "2", "ge", false},
		{"2", "1.2345", "ge", true},
		{"3.1415", "-2", "ge", true},
		{"-3.1415", "-2", "ge", false},
		{"3.1415", "3.1415", "ge", true},
		{"3.14150000000000000000001", "3.1415", "ge", true},

		{"1.0000", "2", "lt", true},
		{"2", "1.2345", "lt", false},
		{"3.1415", "-2", "lt", false},
		{"-3.1415", "-2", "lt", true},
		{"3.1415", "3.1415", "lt", false},
		{"3.1415", "3.14150000000000000000001", "lt", true},

		{"1.0000", "2", "le", true},
		{"2", "1.2345", "le", false},
		{"3.1415", "-2", "le", false},
		{"-3.1415", "-2", "le", true},
		{"3.1415", "3.1415", "le", true},
		{"3.1415", "3.14150000000000000000001", "le", true},
	}

	for i, tt := range tests {
		a, err := NewPrecise(tt.a)
		if err != nil {
			t.Errorf("failed on %d: NewPrecise (a): %v", i, err)
			continue
		}

		b, err := NewPrecise(tt.b)
		if err != nil {
			t.Errorf("failed on %d: NewPrecise (b): %v", i, err)
			continue
		}

		var output bool
		switch tt.op {
		case "gt":
			output = a.GreaterThan(b)
		case "ge":
			output = a.GreaterThanEqual(b)
		case "lt":
			output = a.LessThan(b)
		case "le":
			output = a.LessThanEqual(b)
		default:
			t.Errorf("failed on %d: unknown operator: %s", i, tt.op)
			continue
		}

		if output != tt.output {
			t.Errorf("failed on %d: have %t, want %t", i, output, tt.output)
		}
	}
}

func TestPreciseQuantize(t *testing.T) {
	tests := []struct {
		input     string
		precision int
		output    string
	}{
		{"123.0000987654", 7, "123.0000987"},
	}

	for i, tt := range tests {
		precise, err := NewPrecise(tt.input)
		if err != nil {
			t.Errorf("failed on %d: NewPrecise: %v", i, err)
			continue
		}

		output := precise.Quantize(tt.precision, Truncate).String()
		if output != tt.output {
			t.Errorf("failed on %d: have %s, want %s", i, output, tt.output)
		}
	}
}

func TestPreciseMod(t *testing.T) {
	tests := []struct {
		a      string
		b      string
		output string
	}{
		{"57.123", "10", "7.123"},
		{"18", "6", "0"},
		{"10.1", "0.5", "0.1"},
		{"10000000", "5555", "1000"},
		{"5550", "120", "30"},
	}

	for i, tt := range tests {
		output, err := PreciseStringMod(tt.a, tt.b)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if output != tt.output {
			t.Errorf("failed on %d: have %s, want %s", i, output, tt.output)
		}
	}
}
