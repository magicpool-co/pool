package types

import (
	"context"
	"math/big"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
)

type BlockBuilder interface {
	SerializeHeader(work *StratumWork) ([]byte, []byte, error)
	SerializeBlock(work *StratumWork) ([]byte, error)
	PartialJob() []interface{}
}

type StratumJob struct {
	HostID       string
	ID           string
	Header       *Hash
	HeaderHash   *Hash
	Seed         *Hash
	Height       *Number
	Difficulty   *Difficulty
	Version      *Number
	BlockBuilder BlockBuilder
	CoinbaseTxID *Hash
	Data         interface{} // @TODO: fix this (allow you to store a struct of a lot of data [for AE])
}

type StratumWork struct {
	WorkerID         string
	JobID            string
	Nonce            *Number
	Hash             *Hash
	MixDigest        *Hash     // for ethash/progpow
	CuckooSolution   *Solution // for cuckoo
	EquihashSolution []byte    // for equihash
}

type TxInput struct {
	Index      uint32
	Hash       string
	Value      *big.Int
	FeeBalance *big.Int
	Data       []byte
}

type TxOutput struct {
	Address    string
	Value      *big.Int
	FeeBalance *big.Int
	SplitFee   bool
	Fees       uint64
}

type UTXOResponse struct {
	Hash    string
	Index   uint32
	Value   uint64
	Address string
}

type TxResponse struct {
	Hash        string
	BlockHash   string
	BlockNumber uint64
	From        string
	To          string
	Value       *big.Int
	Fee         *big.Int
	FeeBalance  *big.Int
	Confirmed   bool
	Outputs     []*UTXOResponse
}

type PayoutNode interface {
	Chain() string
	Address() string
	GetUnits() *Number
	ValidateAddress(string) bool

	// tx helpers
	GetBalance(string) (*big.Int, error)
	GetTx(string) (*TxResponse, error)
	CreateTx([]*TxInput, []*TxOutput) (string, error)
	BroadcastTx(string) (string, error)
}

type MiningNode interface {
	PayoutNode
	Mocked() bool

	// constants
	GetShareDifficulty() *Difficulty
	GetAdjustedShareDifficulty() float64
	GetMaxDifficulty() *big.Int
	GetImmatureDepth() uint64
	GetMatureDepth() uint64
	CalculateHashrate(float64, float64) float64

	// stratum helpers
	GetSubscribeResponse([]byte, string, string) (interface{}, error)
	GetDifficultyRequest() (interface{}, error)
	MarshalJob(interface{}, *StratumJob, bool) (interface{}, error)
	ParseWork([]json.RawMessage, string) (*StratumWork, error)

	// mining helpers
	GetStatus() (uint64, bool, error)
	PingHosts() ([]string, []uint64, []bool, []error)
	GetBlocks(uint64, uint64) ([]*tsdb.RawBlock, error)
	JobNotify(context.Context, time.Duration, chan *StratumJob, chan error)
	SubmitWork(*StratumJob, *StratumWork) (ShareStatus, *pooldb.Round, error)
	UnlockRound(*pooldb.Round) error
	GetBlockExplorerURL(*pooldb.Round) string
}
