package protowire

import (
	"fmt"
)

var (
	ErrUnknownMessage        = fmt.Errorf("message could not be casted as *KaspadMessage")
	ErrRouteClosed           = fmt.Errorf("route is closed")
	ErrRouteAtCapacity       = fmt.Errorf("route at capacity")
	ErrRouteTimedOut         = fmt.Errorf("route timed out")
	ErrRouteNotFound         = fmt.Errorf("route not found")
	ErrClientNotConnected    = fmt.Errorf("client not connected")
	ErrRequestTimedOut       = fmt.Errorf("request timed out")
	ErrClientClosedOnRequest = fmt.Errorf("client closed on request")

	CmdSubmitBlock                            = "SubmitBlock"
	CmdGetBlockTemplate                       = "GetBlockTemplate"
	CmdGetSelectedTipHash                     = "GetSelectedTipHash"
	CmdSubmitTransaction                      = "SubmitTransaction"
	CmdGetBlock                               = "GetBlock"
	CmdGetVirtualSelectedParentChainFromBlock = "GetVirtualSelectedParentChainFromBlock"
	CmdGetBlocks                              = "GetBlocks"
	CmdGetUtxosByAddresses                    = "GetUtxosByAddresses"
	CmdGetInfo                                = "GetInfo"
	CmdGetBalanceByAddress                    = "GetBalanceByAddress"
	CmdUnknown                                = "Unknown"

	Cmds = []string{
		CmdSubmitBlock,
		CmdGetBlockTemplate,
		CmdGetSelectedTipHash,
		CmdSubmitTransaction,
		CmdGetBlock,
		CmdGetVirtualSelectedParentChainFromBlock,
		CmdGetBlocks,
		CmdGetUtxosByAddresses,
		CmdGetInfo,
		CmdGetBalanceByAddress,
		CmdUnknown,
	}
)

func (m *KaspadMessage) getCmd() string {
	if m == nil {
		return CmdUnknown
	} else if m.Payload == nil {
		return CmdUnknown
	}

	switch m.Payload.(type) {
	case *KaspadMessage_SubmitBlockRequest, *KaspadMessage_SubmitBlockResponse:
		return CmdSubmitBlock
	case *KaspadMessage_GetBlockTemplateRequest, *KaspadMessage_GetBlockTemplateResponse:
		return CmdGetBlockTemplate
	case *KaspadMessage_GetSelectedTipHashRequest, *KaspadMessage_GetSelectedTipHashResponse:
		return CmdGetSelectedTipHash
	case *KaspadMessage_SubmitTransactionRequest, *KaspadMessage_SubmitTransactionResponse:
		return CmdSubmitTransaction
	case *KaspadMessage_GetBlockRequest, *KaspadMessage_GetBlockResponse:
		return CmdGetBlock
	case *KaspadMessage_GetVirtualSelectedParentChainFromBlockRequest, *KaspadMessage_GetVirtualSelectedParentChainFromBlockResponse:
		return CmdGetVirtualSelectedParentChainFromBlock
	case *KaspadMessage_GetBlocksRequest, *KaspadMessage_GetBlocksResponse:
		return CmdGetBlocks
	case *KaspadMessage_GetUtxosByAddressesRequest, *KaspadMessage_GetUtxosByAddressesResponse:
		return CmdGetUtxosByAddresses
	case *KaspadMessage_GetInfoRequest, *KaspadMessage_GetInfoResponse:
		return CmdGetInfo
	case *KaspadMessage_GetBalanceByAddressRequest, *KaspadMessage_GetBalanceByAddressResponse:
		return CmdGetBalanceByAddress
	}

	return CmdUnknown
}
