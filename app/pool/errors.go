package pool

import (
	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func errInvalidRequest(id json.RawMessage) interface{} {
	err := rpc.NewResponseWithError(id, 1, "invalid request")
	return err
}

func errInvalidAuthRequest(id json.RawMessage) []interface{} {
	err := rpc.NewResponseWithError(id, 2, "invalid authorization request")
	return []interface{}{err}
}

func errInvalidAddressFormatting(id json.RawMessage) []interface{} {
	err := rpc.NewResponseWithError(id, 3, "invalid address:chain formatting")
	return []interface{}{err}
}

func errInvalidChain(id json.RawMessage) []interface{} {
	err := rpc.NewResponseWithError(id, 4, "invalid chain")
	return []interface{}{err}
}

func errInvalidAddress(id json.RawMessage) []interface{} {
	err := rpc.NewResponseWithError(id, 5, "invalid address")
	return []interface{}{err}
}

func errWorkerNameTooLong(id json.RawMessage) []interface{} {
	err := rpc.NewResponseWithError(id, 6, "worker name too long, max 32 characters")
	return []interface{}{err}
}
