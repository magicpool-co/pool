package kas

import (
	"fmt"

	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
)

func (node Node) execAsGRPC(hostID, method string, req interface{}) (*protowire.KaspadMessage, error) {
	var res interface{}
	var err error
	if hostID == "" {
		res, err = node.grpcHost.Exec(req)
	} else {
		res, _, err = node.grpcHost.ExecSticky(hostID, req)
	}
	if err != nil {
		return nil, err
	}

	msg, ok := res.(*protowire.KaspadMessage)
	if !ok {
		return nil, fmt.Errorf("%s: unable to cast as KaspadMessage", method)
	}

	return msg, nil
}

func handleRPCError(method string, err *protowire.RPCError) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %s", method, err.String())
}

func (node Node) getInfo(hostID string) (bool, error) {
	const method = "getInfo"

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetInfoRequest{
			GetInfoRequest: &protowire.GetInfoRequestMessage{},
		},
	}

	res, err := node.execAsGRPC(hostID, method, req)
	if err != nil {
		return false, err
	}

	obj := res.GetGetInfoResponse()
	if err = handleRPCError(method, obj.Error); err != nil {
		return false, err
	}

	return obj.IsSynced, nil
}

func (node Node) GetInfo() (bool, error) {
	return node.getInfo("")
}

func (node Node) getSelectedTipHash(hostID string) (string, error) {
	const method = "getSelectedTipHash"

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetSelectedTipHashRequest{
			GetSelectedTipHashRequest: &protowire.GetSelectedTipHashRequestMessage{},
		},
	}

	res, err := node.execAsGRPC(hostID, method, req)
	if err != nil {
		return "", err
	}

	obj := res.GetGetSelectedTipHashResponse()
	if err = handleRPCError(method, obj.Error); err != nil {
		return "", err
	} else if obj.SelectedTipHash == "" {
		return "", fmt.Errorf("unable to find selected tip hash")
	}

	return obj.SelectedTipHash, nil
}

func (node Node) GetTip() (string, error) {
	return node.getSelectedTipHash("")
}

func (node Node) getBlock(hostID, hash string) (*Block, error) {
	const method = "getBlock"

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetBlockRequest{
			GetBlockRequest: &protowire.GetBlockRequestMessage{
				Hash:                hash,
				IncludeTransactions: true,
			},
		},
	}

	res, err := node.execAsGRPC(hostID, method, req)
	if err != nil {
		return nil, err
	}

	obj := res.GetGetBlockResponse()
	if err = handleRPCError(method, obj.Error); err != nil {
		return nil, err
	} else if obj.Block == nil {
		return nil, fmt.Errorf("unable to find block: %s", hash)
	}

	return nil, nil
}

func (node Node) GetBlock(hash string) (*Block, error) {
	return node.getBlock("", hash)
}

func (node Node) getBlockTemplate(extraData string) (*Block, error) {
	const method = "getBlockTemplate"

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_GetBlockTemplateRequest{
			GetBlockTemplateRequest: &protowire.GetBlockTemplateRequestMessage{
				PayAddress: node.address,
				ExtraData:  extraData,
			},
		},
	}

	res, err := node.execAsGRPC("", method, req)
	if err != nil {
		return nil, err
	}

	obj := res.GetGetBlockTemplateResponse()
	if err = handleRPCError(method, obj.Error); err != nil {
		return nil, err
	} else if !obj.IsSynced {
		return nil, fmt.Errorf("node is not synced")
	} else if obj.Block == nil {
		return nil, fmt.Errorf("unable to find block template")
	}

	fmt.Println(obj.Block.Header)

	return nil, nil
}

func (node Node) GetBlockTemplate() (*Block, error) {
	return node.getBlockTemplate("")
}

func (node Node) submitBlock(block *Block) error {
	const method = "submitBlock"

	parents := make([]*protowire.RpcBlockLevelParents, len(block.Parents))
	for i, hashes := range block.Parents {
		parents[i] = &protowire.RpcBlockLevelParents{
			ParentHashes: hashes,
		}
	}

	txs := make([]*protowire.RpcTransaction, len(block.Transactions))
	for i, tx := range txs {
		inputs := make([]*protowire.RpcTransactionInput, len(tx.Inputs))
		for j, input := range tx.Inputs {
			inputs[j] = &protowire.RpcTransactionInput{
				PreviousOutpoint: &protowire.RpcOutpoint{
					TransactionId: input.PreviousOutpoint.TransactionId,
					Index:         input.PreviousOutpoint.Index,
				},
				SignatureScript: input.SignatureScript,
				Sequence:        input.Sequence,
				SigOpCount:      input.SigOpCount,
			}
		}

		outputs := make([]*protowire.RpcTransactionOutput, len(tx.Outputs))
		for j, output := range tx.Outputs {
			outputs[j] = &protowire.RpcTransactionOutput{
				Amount: output.Amount,
				ScriptPublicKey: &protowire.RpcScriptPublicKey{
					Version:         output.ScriptPublicKey.Version,
					ScriptPublicKey: output.ScriptPublicKey.ScriptPublicKey,
				},
			}
		}

		txs[i] = &protowire.RpcTransaction{
			Version:      tx.Version,
			Inputs:       inputs,
			Outputs:      outputs,
			LockTime:     tx.LockTime,
			SubnetworkId: tx.SubnetworkId,
			Gas:          tx.Gas,
			Payload:      tx.Payload,
		}
	}

	rpcBlock := &protowire.RpcBlock{
		Header: &protowire.RpcBlockHeader{
			Version:              block.Version,
			Parents:              parents,
			HashMerkleRoot:       block.HashMerkleRoot,
			AcceptedIdMerkleRoot: block.AcceptedIdMerkleRoot,
			UtxoCommitment:       block.UtxoCommitment,
			Timestamp:            block.Timestamp,
			Bits:                 block.Bits,
			Nonce:                block.Nonce,
			DaaScore:             block.DaaScore,
			BlueWork:             block.BlueWork,
			PruningPoint:         block.PruningPoint,
			BlueScore:            block.BlueScore,
		},
		Transactions: txs,
	}

	req := &protowire.KaspadMessage{
		Payload: &protowire.KaspadMessage_SubmitBlockRequest{
			SubmitBlockRequest: &protowire.SubmitBlockRequestMessage{
				Block:             rpcBlock,
				AllowNonDAABlocks: false,
			},
		},
	}

	res, err := node.execAsGRPC("", method, req)
	if err != nil {
		return err
	}

	obj := res.GetSubmitBlockResponse()
	if err = handleRPCError(method, obj.Error); err != nil {
		return fmt.Errorf("%v: %s", err, obj.RejectReason.String())
	} else if obj.RejectReason != 0 {
		return fmt.Errorf("rejected block: %s", obj.RejectReason.String())
	}

	return nil
}
