package types

import (
	"encoding/json"
	"testing"
)

func mustMarshalJSON(input interface{}) []byte {
	data, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}

	return data
}

func TestHashUnmarshal(t *testing.T) {
	tests := []struct {
		input     []byte
		outputHex string
	}{
		{
			input:     mustMarshalJSON(`2c32`),
			outputHex: "2c32",
		},
		{
			input:     mustMarshalJSON(`0x2c32`),
			outputHex: "2c32",
		},
	}

	for i, tt := range tests {
		var hash *Hash
		err := json.Unmarshal(tt.input, &hash)
		if err != nil {
			t.Errorf("failed on %d: unmarshal output: %v", i, err)
		} else if hash.Hex() != tt.outputHex {
			t.Errorf("failed on %d: output mismatch: have %s, want %s", i, hash.Hex(), tt.outputHex)
		}
	}
}

func TestNumberUnmarshal(t *testing.T) {
	tests := []struct {
		input       []byte
		outputValue uint64
	}{
		{
			input:       mustMarshalJSON(0x2c32),
			outputValue: 0x2c32,
		},
		{
			input:       mustMarshalJSON(6161047830682206209),
			outputValue: 6161047830682206209,
		},
		{
			input:       mustMarshalJSON(`2c32`),
			outputValue: 0x2c32,
		},
		{
			input:       mustMarshalJSON(`0x2c32`),
			outputValue: 0x2c32,
		},
	}

	for i, tt := range tests {
		var number *Number
		err := json.Unmarshal(tt.input, &number)
		if err != nil {
			t.Errorf("failed on %d: unmarshal output: %v", i, err)
		} else if number.Value() != tt.outputValue {
			t.Errorf("failed on %d: output mismatch: have %d, want %d", i, number.Value(), tt.outputValue)
		}
	}
}

func TestSolutionUnmarshal(t *testing.T) {
	tests := []struct {
		input     []byte
		outputHex string
	}{
		{
			input:     mustMarshalJSON("12321232"),
			outputHex: "12321232",
		},
		{
			input:     mustMarshalJSON([]string{"0x12", "0x32"}),
			outputHex: "0000001200000032",
		},
		{
			input:     mustMarshalJSON([]uint64{0x12, 0x32}),
			outputHex: "0000001200000032",
		},
	}

	for i, tt := range tests {
		var solution *Solution
		err := json.Unmarshal(tt.input, &solution)
		if err != nil {
			t.Errorf("failed on %d: unmarshal output: %v", i, err)
		} else if solution.Hex() != tt.outputHex {
			t.Errorf("failed on %d: output mismatch: have %s, want %s", i, solution.Hex(), tt.outputHex)
		}
	}
}
