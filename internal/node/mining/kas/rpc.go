package kas

import (
	"fmt"

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

func (node Node) getInfo(hostID string) (bool, error) {
	const method = "getInfo"

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetInfoRequest{
			GetInfoRequest: &protowire.GetInfoRequestMessage{},
		},
	}

	res, _, err := node.execAsGRPCSticky(hostID, method, req)
	if err != nil {
		return false, err
	}

	obj := res.GetGetInfoResponse()
	if obj == nil {
		return false, fmt.Errorf("empty info response")
	} else if err = handleRPCError(method, obj.Error); err != nil {
		return false, err
	}

	return obj.IsSynced, nil
}

func (node Node) getSelectedTipHash(hostID string) (string, error) {
	const method = "getSelectedTipHash"

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetSelectedTipHashRequest{
			GetSelectedTipHashRequest: &protowire.GetSelectedTipHashRequestMessage{},
		},
	}

	res, _, err := node.execAsGRPCSticky(hostID, method, req)
	if err != nil {
		return "", err
	}

	obj := res.GetGetSelectedTipHashResponse()
	if obj == nil {
		return "", fmt.Errorf("empty tip response")
	} else if err = handleRPCError(method, obj.Error); err != nil {
		return "", err
	} else if obj.SelectedTipHash == "" {
		return "", fmt.Errorf("unable to find selected tip hash")
	}

	return obj.SelectedTipHash, nil
}

func (node Node) getBlock(hostID, hash string, includeTxs bool) (*Block, error) {
	const method = "getBlock"

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetBlockRequest{
			GetBlockRequest: &protowire.GetBlockRequestMessage{
				Hash:                hash,
				IncludeTransactions: includeTxs,
			},
		},
	}

	res, _, err := node.execAsGRPCSticky(hostID, method, req)
	if err != nil {
		return nil, err
	}

	obj := res.GetGetBlockResponse()
	if obj == nil {
		return nil, fmt.Errorf("empty response for block: %s", hash)
	} else if err = handleRPCError(method, obj.Error); err != nil {
		return nil, err
	} else if obj.Block == nil {
		return nil, fmt.Errorf("unable to find block: %s", hash)
	}

	return protowireToBlock(obj.Block), nil
}

func (node Node) getBlockTemplate(extraData string) (*Block, string, error) {
	const method = "getBlockTemplate"

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetBlockTemplateRequest{
			GetBlockTemplateRequest: &protowire.GetBlockTemplateRequestMessage{
				PayAddress: node.address,
				ExtraData:  extraData,
			},
		},
	}

	res, hostID, err := node.execAsGRPCSticky("", method, req)
	if err != nil {
		return nil, hostID, err
	}

	obj := res.GetGetBlockTemplateResponse()
	if obj == nil {
		return nil, hostID, fmt.Errorf("empty template response")
	} else if err = handleRPCError(method, obj.Error); err != nil {
		return nil, hostID, err
	} else if !obj.IsSynced {
		return nil, hostID, fmt.Errorf("node is not synced")
	} else if obj.Block == nil {
		return nil, hostID, fmt.Errorf("unable to find block template")
	}

	return protowireToBlock(obj.Block), hostID, nil
}

func (node Node) submitBlock(hostID string, block *Block) error {
	const method = "submitBlock"

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
		return fmt.Errorf("empty submit response")
	} else if err = handleRPCError(method, obj.Error); err != nil {
		return fmt.Errorf("%v: %s", err, obj.RejectReason.String())
	} else if obj.RejectReason != 0 {
		return fmt.Errorf("rejected block: %s", obj.RejectReason.String())
	}

	return nil
}
