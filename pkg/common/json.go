package common

import (
	"fmt"

	"github.com/goccy/go-json"
)

var (
	JsonTrue  = MustMarshalJSON(true)
	JsonFalse = MustMarshalJSON(false)
	JsonZero  = MustMarshalJSON(0)
)

func MustMarshalJSON(inp interface{}) []byte {
	data, err := json.Marshal(inp)
	if err != nil {
		panic(fmt.Errorf("MustMarshalJSON: %v", err))
	}

	return data
}
