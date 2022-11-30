package kas

import (
	"context"
	"time"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sencha-dev/powkit/heavyhash"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
)

var (
	mainnetPrefix = "kaspa"
	testnetPrefix = "kaspatest"
)

func generateHost(urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*hostpool.GRPCPool, error) {
	var (
		port            = 16110
		hostHealthCheck = &hostpool.GRPCHealthCheck{
			Request: &protowire.KaspadMessage{
				Payload: &protowire.KaspadMessage_GetSelectedTipHashRequest{
					GetSelectedTipHashRequest: &protowire.GetSelectedTipHashRequestMessage{},
				},
			},
		}
	)

	if len(urls) == 0 {
		return nil, nil
	}

	factory := func(url string, timeout time.Duration) (hostpool.GRPCClient, error) {
		return protowire.NewClient(url, timeout)
	}

	host := hostpool.NewGRPCPool(context.Background(), factory, logger, hostHealthCheck, tunnel)
	for _, url := range urls {
		err := host.AddHost(url, port)
		if err != nil {
			return nil, err
		}
	}

	return host, nil
}

func New(mainnet bool, urls []string, rawPriv string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*Node, error) {
	prefix := mainnetPrefix
	if !mainnet {
		prefix = testnetPrefix
	}

	grpcHost, err := generateHost(urls, logger, tunnel)
	if err != nil {
		return nil, err
	}

	obscuredPriv, err := crypto.ObscureHex(rawPriv)
	if err != nil {
		return nil, err
	} else if err := crypto.ValidateSecp256k1PrivateKey(obscuredPriv); err != nil {
		return nil, err
	}

	privKey := secp256k1.PrivKeyFromBytes(obscuredPriv)
	pubKeyBytes := privKey.PubKey().SerializeCompressed()
	address, err := bech32.EncodeBCH(addressCharset, prefix, pubKeyECDSAAddrID, pubKeyBytes)
	if err != nil {
		return nil, err
	}

	node := &Node{
		mocked:   grpcHost == nil,
		mainnet:  mainnet,
		prefix:   prefix,
		address:  address,
		privKey:  privKey,
		grpcHost: grpcHost,
		pow:      heavyhash.NewKaspa(),
		logger:   logger,
	}

	return node, nil
}

type Node struct {
	mocked   bool
	mainnet  bool
	prefix   string
	address  string
	privKey  *secp256k1.PrivateKey
	grpcHost *hostpool.GRPCPool
	pow      *heavyhash.Client
	logger   *log.Logger
}

type TransactionOutpoint struct {
	TransactionId string
	Index         uint32
}

type TransactionInput struct {
	PreviousOutpoint *TransactionOutpoint
	SignatureScript  string
	Sequence         uint64
	SigOpCount       uint32
}

type TransactionScriptPubKey struct {
	Version         uint32
	ScriptPublicKey string
}

type TransactionOutput struct {
	Amount          uint64
	ScriptPublicKey *TransactionScriptPubKey
}

type Transaction struct {
	Version      uint32
	Inputs       []*TransactionInput
	Outputs      []*TransactionOutput
	LockTime     uint64
	SubnetworkId string
	Gas          uint64
	Payload      string
}

func protowireToTransaction(rpcTx *protowire.RpcTransaction) *Transaction {
	inputs := make([]*TransactionInput, len(rpcTx.Inputs))
	for j, input := range rpcTx.Inputs {
		inputs[j] = &TransactionInput{
			PreviousOutpoint: &TransactionOutpoint{
				TransactionId: input.PreviousOutpoint.TransactionId,
				Index:         input.PreviousOutpoint.Index,
			},
			SignatureScript: input.SignatureScript,
			Sequence:        input.Sequence,
			SigOpCount:      input.SigOpCount,
		}
	}

	outputs := make([]*TransactionOutput, len(rpcTx.Outputs))
	for j, output := range rpcTx.Outputs {
		outputs[j] = &TransactionOutput{
			Amount: output.Amount,
			ScriptPublicKey: &TransactionScriptPubKey{
				Version:         output.ScriptPublicKey.Version,
				ScriptPublicKey: output.ScriptPublicKey.ScriptPublicKey,
			},
		}
	}

	tx := &Transaction{
		Version:      rpcTx.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     rpcTx.LockTime,
		SubnetworkId: rpcTx.SubnetworkId,
		Gas:          rpcTx.Gas,
		Payload:      rpcTx.Payload,
	}

	return tx
}

func transactionToProtowire(tx *Transaction) *protowire.RpcTransaction {
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

	rpcTx := &protowire.RpcTransaction{
		Version:      tx.Version,
		Inputs:       inputs,
		Outputs:      outputs,
		LockTime:     tx.LockTime,
		SubnetworkId: tx.SubnetworkId,
		Gas:          tx.Gas,
		Payload:      tx.Payload,
	}

	return rpcTx
}

type Block struct {
	// Header
	Version              uint32
	Parents              [][]string
	HashMerkleRoot       string
	AcceptedIdMerkleRoot string
	UtxoCommitment       string
	Timestamp            int64
	Bits                 uint32
	Nonce                uint64
	DaaScore             uint64
	BlueWork             string
	PruningPoint         string
	BlueScore            uint64
	// Transactions
	Transactions []*Transaction
	// VerboseData
	Hash                string
	Difficulty          float64
	MergeSetBluesHashes []string
	MergeSetRedsHashes  []string
	ChildrenHashes      []string
	IsChainBlock        bool
}

func protowireToBlock(rpcBlock *protowire.RpcBlock) *Block {
	parents := make([][]string, len(rpcBlock.Header.Parents))
	for i, parent := range rpcBlock.Header.Parents {
		parents[i] = make([]string, len(parent.ParentHashes))
		for j, hash := range parent.ParentHashes {
			parents[i][j] = hash
		}
	}

	txs := make([]*Transaction, len(rpcBlock.Transactions))
	for i, rpcTx := range rpcBlock.Transactions {
		txs[i] = protowireToTransaction(rpcTx)
	}

	block := &Block{
		Version:              rpcBlock.Header.Version,
		Parents:              parents,
		HashMerkleRoot:       rpcBlock.Header.HashMerkleRoot,
		AcceptedIdMerkleRoot: rpcBlock.Header.AcceptedIdMerkleRoot,
		UtxoCommitment:       rpcBlock.Header.UtxoCommitment,
		Timestamp:            rpcBlock.Header.Timestamp,
		Bits:                 rpcBlock.Header.Bits,
		Nonce:                rpcBlock.Header.Nonce,
		DaaScore:             rpcBlock.Header.DaaScore,
		BlueWork:             rpcBlock.Header.BlueWork,
		PruningPoint:         rpcBlock.Header.PruningPoint,
		BlueScore:            rpcBlock.Header.BlueScore,
		Transactions:         txs,
		Hash:                 rpcBlock.VerboseData.Hash,
		Difficulty:           rpcBlock.VerboseData.Difficulty,
		MergeSetBluesHashes:  rpcBlock.VerboseData.MergeSetBluesHashes,
		MergeSetRedsHashes:   rpcBlock.VerboseData.MergeSetRedsHashes,
		ChildrenHashes:       rpcBlock.VerboseData.ChildrenHashes,
		IsChainBlock:         rpcBlock.VerboseData.IsChainBlock,
	}

	return block
}

func blockToProtowire(block *Block) *protowire.RpcBlock {
	parents := make([]*protowire.RpcBlockLevelParents, len(block.Parents))
	for i, hashes := range block.Parents {
		parents[i] = &protowire.RpcBlockLevelParents{
			ParentHashes: hashes,
		}
	}

	rpcTxs := make([]*protowire.RpcTransaction, len(block.Transactions))
	for i, tx := range block.Transactions {
		rpcTxs[i] = transactionToProtowire(tx)
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
		Transactions: rpcTxs,
	}

	return rpcBlock
}
