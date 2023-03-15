package types

import (
	"context"
	"math/big"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
)

/* stratum */

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
	Timestamp    time.Time
	Version      *Number
	BlockBuilder BlockBuilder
	CoinbaseTxID *Hash
	Data         interface{} // @TODO: fix this (allow you to store a struct of a lot of data [for AE/KAS])
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

/* tx */

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
	Fee        *big.Int
	FeeBalance *big.Int
	SplitFee   bool
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

/* node */

type PayoutNode interface {
	Name() string
	Chain() string
	Address() string
	GetUnits() *Number
	GetAccountingType() AccountingType
	GetAddressPrefix() string
	ValidateAddress(string) bool

	// tx helpers
	GetTxExplorerURL(string) string
	GetAddressExplorerURL(string) string
	GetBalance() (*big.Int, error)
	GetTx(string) (*TxResponse, error)
	CreateTx([]*TxInput, []*TxOutput) (string, string, error)
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
	GetSubscribeResponses([]byte, string, string) ([]interface{}, error)
	GetAuthorizeResponses() ([]interface{}, error)
	GetClientType(string) int
	MarshalJob(interface{}, *StratumJob, bool, int) (interface{}, error)
	ParseWork([]json.RawMessage, string) (*StratumWork, error)

	// mining helpers
	GetBlockExplorerURL(*pooldb.Round) string
	GetStatus() (uint64, bool, error)
	PingHosts() ([]string, []uint64, []bool, []error)
	GetBlocks(uint64, uint64) ([]*tsdb.RawBlock, error)
	GetBlocksByHash(string, uint64) ([]*tsdb.RawBlock, error)
	JobNotify(context.Context, time.Duration) chan *StratumJob
	SubmitWork(*StratumJob, *StratumWork) (ShareStatus, *Hash, *pooldb.Round, error)
	UnlockRound(*pooldb.Round) error
	MatureRound(*pooldb.Round) ([]*pooldb.UTXO, error)
}

/* exchange */

type Exchange interface {
	ID() ExchangeID
	GetTradeTimeout() time.Duration

	// account
	GetAccountStatus() error

	// rate
	GetRate(string) (float64, error)
	GetHistoricalRates(string, time.Time, time.Time, bool) (map[time.Time]float64, error)
	GetOutputThresholds() map[string]*big.Int
	GetPrices(map[string]map[string]*big.Int) (map[string]map[string]float64, error)

	// wallet
	GetWalletStatus(string) (bool, bool, error)
	GetWalletBalance(string) (float64, float64, error)

	// deposit
	GetDepositAddress(string) (string, error)
	GetDepositByTxID(string, string) (*Deposit, error)
	GetDepositByID(string, string) (*Deposit, error)

	// transfer
	TransferToTradeAccount(string, float64) error
	TransferToMainAccount(string, float64) error

	// trade
	GenerateTradePath(string, string) ([]*Trade, error)
	CreateTrade(string, TradeDirection, float64) (string, error)
	GetTradeByID(string, string, float64) (*Trade, error)
	CancelTradeByID(string, string) error

	// withdrawal
	CreateWithdrawal(string, string, float64) (string, error)
	GetWithdrawalByID(string, string) (*Withdrawal, error)
	NeedsWithdrawalFeeSubtraction() bool
}

type Deposit struct {
	ID        string
	TxID      string
	Value     string
	Fee       string
	Completed bool
}

type Market struct {
	Market    string
	Base      string
	Quote     string
	Direction TradeDirection
}

type Trade struct {
	ID        string
	FromChain string
	ToChain   string
	Market    string
	Direction TradeDirection
	Increment int

	Value    string
	Proceeds string
	Fees     string
	Price    string

	Completed bool
	Active    bool
}

type Withdrawal struct {
	ID        string
	TxID      string
	Value     string
	Fee       string
	Completed bool
}
