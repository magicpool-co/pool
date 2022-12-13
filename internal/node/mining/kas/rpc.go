package kas

import (
	"fmt"

	"github.com/magicpool-co/pool/internal/node/mining/kas/mock"
	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
)

func handleRPCError(method string, err *protowire.RPCError) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %s", method, err.String())
}

func (node Node) execAsGRPC(method string, req interface{}) (*protowire.KaspadMessage, error) {
	res, err := node.grpcHost.Exec(req)
	if err != nil {
		return nil, err
	}

	msg, ok := res.(*protowire.KaspadMessage)
	if !ok {
		return nil, fmt.Errorf("%s: unable to cast as KaspadMessage", method)
	}

	return msg, nil
}

func (node Node) execAsGRPCSticky(hostID, method string, req interface{}) (*protowire.KaspadMessage, string, error) {
	res, hostID, err := node.grpcHost.ExecSticky(hostID, req)
	if err != nil {
		return nil, "", err
	}

	msg, ok := res.(*protowire.KaspadMessage)
	if !ok {
		return nil, "", fmt.Errorf("%s: unable to cast as KaspadMessage", method)
	}

	return msg, hostID, nil
}

func (node Node) execAsGRPCSynced(method string, req interface{}) (*protowire.KaspadMessage, string, error) {
	res, hostID, err := node.grpcHost.ExecSynced(req)
	if err != nil {
		return nil, "", err
	}

	msg, ok := res.(*protowire.KaspadMessage)
	if !ok {
		return nil, "", fmt.Errorf("%s: unable to cast as KaspadMessage", method)
	}

	return msg, hostID, nil
}

func (node Node) getInfo(hostID string) (bool, error) {
	const method = "getInfo"

	var obj *protowire.GetInfoResponseMessage
	if node.mocked {
		obj = mock.GetInfo()
	} else {
		req := &protowire.KaspadMessage{
			Payload: &protowire.KaspadMessage_GetInfoRequest{
				GetInfoRequest: &protowire.GetInfoRequestMessage{},
			},
		}

		res, _, err := node.execAsGRPCSticky(hostID, method, req)
		if err != nil {
			return false, err
		}

		obj = res.GetGetInfoResponse()
		if obj == nil {
			return false, fmt.Errorf("empty info response")
		} else if err = handleRPCError(method, obj.Error); err != nil {
			return false, err
		}
	}

	return obj.IsSynced, nil
}

func (node Node) getSelectedTipHash(hostID string) (string, error) {
	const method = "getSelectedTipHash"

	var obj *protowire.GetSelectedTipHashResponseMessage
	if node.mocked {
		obj = mock.GetSelectedTipHash()
	} else {
		req := &protowire.KaspadMessage{
			Payload: &protowire.KaspadMessage_GetSelectedTipHashRequest{
				GetSelectedTipHashRequest: &protowire.GetSelectedTipHashRequestMessage{},
			},
		}

		var res *protowire.KaspadMessage
		var err error
		if hostID == "" {
			res, _, err = node.execAsGRPCSynced(method, req)
		} else {
			res, _, err = node.execAsGRPCSticky(hostID, method, req)
		}

		if err != nil {
			return "", err
		}

		obj = res.GetGetSelectedTipHashResponse()
		if obj == nil {
			return "", fmt.Errorf("empty tip response")
		} else if err = handleRPCError(method, obj.Error); err != nil {
			return "", err
		} else if obj.SelectedTipHash == "" {
			return "", fmt.Errorf("unable to find selected tip hash")
		}
	}

	return obj.SelectedTipHash, nil
}

func (node Node) getBlock(hostID, hash string, includeTxs bool) (*Block, error) {
	const method = "getBlock"

	var obj *protowire.GetBlockResponseMessage
	if node.mocked {
		obj = mock.GetBlock()
	} else {
		req := &protowire.KaspadMessage{
			Payload: &protowire.KaspadMessage_GetBlockRequest{
				GetBlockRequest: &protowire.GetBlockRequestMessage{
					Hash:                hash,
					IncludeTransactions: includeTxs,
				},
			},
		}

		var res *protowire.KaspadMessage
		var err error
		if hostID == "" {
			res, _, err = node.execAsGRPCSynced(method, req)
		} else {
			res, _, err = node.execAsGRPCSticky(hostID, method, req)
		}

		obj = res.GetGetBlockResponse()
		if obj == nil {
			return nil, fmt.Errorf("empty response for block: %s", hash)
		} else if err = handleRPCError(method, obj.Error); err != nil {
			return nil, err
		} else if obj.Block == nil {
			return nil, fmt.Errorf("unable to find block: %s", hash)
		}
	}

	return protowireToBlock(obj.Block), nil
}

func (node Node) getBlockTemplate(extraData string) (*Block, string, error) {
	const method = "getBlockTemplate"

	var obj *protowire.GetBlockTemplateResponseMessage
	var hostID string
	if node.mocked {
		obj = mock.GetBlockTemplate()
	} else {
		req := &protowire.KaspadMessage{
			Payload: &protowire.KaspadMessage_GetBlockTemplateRequest{
				GetBlockTemplateRequest: &protowire.GetBlockTemplateRequestMessage{
					PayAddress: node.address,
					ExtraData:  extraData,
				},
			},
		}

		res, rawHostID, err := node.execAsGRPCSynced(method, req)
		hostID = rawHostID
		if err != nil {
			return nil, hostID, err
		}

		obj = res.GetGetBlockTemplateResponse()
		if obj == nil {
			return nil, hostID, fmt.Errorf("empty template response")
		} else if err = handleRPCError(method, obj.Error); err != nil {
			return nil, hostID, err
		} else if !obj.IsSynced {
			node.grpcHost.SetHostSyncStatus(hostID, false)
			return nil, hostID, fmt.Errorf("node is not synced")
		} else if obj.Block == nil {
			return nil, hostID, fmt.Errorf("unable to find block template")
		}
	}

	return protowireToBlock(obj.Block), hostID, nil
}

func (node Node) submitBlock(hostID string, block *Block) error {
	const method = "submitBlock"

	if !node.mocked {
		req := &protowire.KaspadMessage{
			Payload: &protowire.KaspadMessage_SubmitBlockRequest{
				SubmitBlockRequest: &protowire.SubmitBlockRequestMessage{
					Block:             blockToProtowire(block),
					AllowNonDAABlocks: false,
				},
			},
		}

		res, _, err := node.execAsGRPCSticky(hostID, method, req)
		if err != nil {
			return err
		}

		obj := res.GetSubmitBlockResponse()
		if obj == nil {
			return fmt.Errorf("empty submit block response")
		} else if err = handleRPCError(method, obj.Error); err != nil {
			return fmt.Errorf("%v: %s", err, obj.RejectReason.String())
		} else if obj.RejectReason != 0 {
			return fmt.Errorf("rejected block: %s", obj.RejectReason.String())
		}
	}

	return nil
}

func (node Node) getBalanceByAddress(address string) (uint64, error) {
	const method = "getBalanceByAddress"

	if node.mocked {
		return 0, nil
	}

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetBalanceByAddressRequest{
			GetBalanceByAddressRequest: &protowire.GetBalanceByAddressRequestMessage{
				Address: address,
			},
		},
	}

	res, _, err := node.execAsGRPCSynced(method, req)
	if err != nil {
		return 0, err
	}

	obj := res.GetGetBalanceByAddressResponse()
	if obj == nil {
		return 0, nil
	} else if err = handleRPCError(method, obj.Error); err != nil {
		return 0, err
	}

	return obj.Balance, nil
}

func (node Node) GetUtxosByAddress(address string) ([]*protowire.UtxosByAddressesEntry, error) {
	const method = "getUtxosByAddresses"

	if node.mocked {
		return nil, nil
	}

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetUtxosByAddressesRequest{
			GetUtxosByAddressesRequest: &protowire.GetUtxosByAddressesRequestMessage{
				Addresses: []string{address},
			},
		},
	}

	res, _, err := node.execAsGRPCSynced(method, req)
	if err != nil {
		return nil, err
	}

	obj := res.GetGetUtxosByAddressesResponse()
	if obj == nil {
		return nil, nil
	} else if err = handleRPCError(method, obj.Error); err != nil {
		return nil, err
	}

	return obj.Entries, nil
}

func (node Node) submitTransaction(tx *protowire.RpcTransaction) (string, error) {
	const method = "submitTransaction"

	if node.mocked {
		return "", nil
	}

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_SubmitTransactionRequest{
			SubmitTransactionRequest: &protowire.SubmitTransactionRequestMessage{
				Transaction: tx,
				AllowOrphan: false,
			},
		},
	}

	res, _, err := node.execAsGRPCSynced(method, req)
	if err != nil {
		return "", err
	}

	obj := res.GetSubmitTransactionResponse()
	if obj == nil {
		return "", fmt.Errorf("empty submit transaction response")
	} else if err = handleRPCError(method, obj.Error); err != nil {
		return "", err
	}

	return obj.TransactionId, nil
}
