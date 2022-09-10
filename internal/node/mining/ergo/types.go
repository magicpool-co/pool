package ergo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brianium/mnemonic"
	"github.com/sencha-dev/powkit/autolykos2"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
)

var (
	addressPrefix = []byte{0x01}
)

func generateHost(urls []string, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, error) {
	var (
		port        = 9053
		hostOptions = &hostpool.HTTPHostOptions{
			Headers: map[string]string{"api_key": "rpcrpc"},
		}
		hostHealthCheck = &hostpool.HTTPHealthCheck{
			HTTPMethod: "GET",
			HTTPPath:   "/info",
		}
	)

	if len(urls) == 0 {
		return nil, nil
	}

	host := hostpool.NewHTTPPool(context.Background(), hostHealthCheck, tunnel)
	for _, url := range urls {
		err := host.AddHost(url, port, hostOptions)
		if err != nil {
			return nil, err
		}
	}

	return host, nil
}

func (node Node) initWallets() error {
	hostIDs := node.httpHost.GetAllHosts()
	for _, hostID := range hostIDs {
		status, err := node.getWalletStatus(hostID)
		if err != nil {
			return err
		} else if !status.IsInitialized {
			err = node.postWalletRestore(hostID)
			if err != nil {
				return err
			}
		}

		if !status.IsUnlocked {
			err = node.postWalletUnlock(hostID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func New(mainnet bool, urls []string, rawPriv string, tunnel *sshtunnel.SSHTunnel) (*Node, error) {
	httpHost, err := generateHost(urls, tunnel)
	if err != nil {
		return nil, err
	}

	obscuredPriv, err := crypto.ObscureHex(rawPriv)
	if err != nil {
		return nil, err
	} else if len(obscuredPriv) < 20 {
		return nil, fmt.Errorf("obscured priv too short")
	}

	mnemonicPhrase, err := mnemonic.New(obscuredPriv[:20], mnemonic.English)
	if err != nil {
		return nil, err
	}

	node := &Node{
		mocked:   httpHost == nil,
		mainnet:  mainnet,
		mnemonic: mnemonicPhrase.Sentence(),
		httpHost: httpHost,
		pow:      autolykos2.NewErgo(),
	}

	node.address, err = node.getRewardAddress()
	if err != nil {
		return nil, err
	}

	if !node.mocked {
		err = node.initWallets()
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

type Node struct {
	mocked   bool
	mainnet  bool
	address  string
	mnemonic string
	httpHost *hostpool.HTTPPool
	pow      *autolykos2.Client
}

type NodeInfo struct {
	Name          string `json:"name"`
	AppVersion    string `json:"appVersion"`
	FullHeight    uint64 `json:"fullHeight"`
	HeadersHeight uint64 `json:"headersHeight"`
	MaxPeerHeight uint64 `json:"maxPeerHeight"`
	Parameters    struct {
		BlockVersion uint64 `json:"blockVersion"`
	} `json:"parameters"`
}

type NodeError struct {
	Error  int    `json:"error"`
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

type WalletStatus struct {
	IsInitialized bool   `json:"isInitialized"`
	IsUnlocked    bool   `json:"isUnlocked"`
	ChangeAddress string `json:"changeAddress`
	WalletHeight  uint64 `json:"walletHeight"`
	Error         string `json:"error"`
}

type PowSolution struct {
	PK string          `json:"pk"`
	W  string          `json:"w"`
	N  string          `json:"n"`
	D  json.RawMessage `json:"d"`
}

type SpendingProof struct {
	ProofBytes string            `json:"proofBytes"`
	Extension  map[string]string `json:"extension"`
}

type TransactionInput struct {
	BoxID         string            `json:"boxId"`
	SpendingProof *SpendingProof    `json:"spendingProof"`
	Extension     map[string]string `json:"extension"`
}

type TransactionAsset struct {
	TokenID string `json:"tokenId"`
	Amount  uint64 `json:"amount"`
}

type TransactionOutput struct {
	BoxID          string `json:"boxId"`
	Value          uint64 `json:"value"`
	ErgoTree       string `json:"ergoTree"`
	CreationHeight uint64 `json:"creationHeight"`
	TransactionsID string `json:"transactionsId"`
	Index          uint64 `json:"index"`
}

type Transaction struct {
	ID      string               `json:"id"`
	Inputs  []*TransactionInput  `json:"inputs"`
	Outputs []*TransactionOutput `json:"outputs"`
	Size    uint64               `json:"size"`
}

type BlockTransactions struct {
	HeaderID     string         `json:"headerId"`
	Transactions []*Transaction `json:"transactions"`
	BlockVersion int            `json:"blockVersion"`
	Size         uint64         `json:"size"`
}

type BlockHeader struct {
	ID               string       `json:"id"`
	Timestamp        int64        `json:"timestamp"`
	Version          int          `json:"version"`
	ADProofsRoot     string       `json:"adProofsRoot"`
	StateRoot        string       `json:"stateRoot"`
	TransactionsRoot string       `json:"transactionsRoot"`
	NBits            uint32       `json:"nBits"`
	ExtensionHash    string       `json:"extensionHash"`
	PowSolutions     *PowSolution `json:"powSolutions"`
	Height           uint64       `json:"height"`
	Difficulty       string       `json:"difficulty"`
	ParentID         string       `json:"parentId"`
	Votes            string       `json:"votes"`
	Size             uint64       `json:"size"`
	ExtensionID      string       `json:"extensionId"`
	TransactionsID   string       `json:"transactionsId"`
	ADProofsID       string       `json:"adProofsId"`
}

type Block struct {
	Header            *BlockHeader       `json:"header"`
	BlockTransactions *BlockTransactions `json:"blockTransactions"`
}

type Balance struct {
	Height  uint64 `json:"height"`
	Balance uint64 `json:"balance"`
}

type Address struct {
	Address string `json:"address"`
}

type RewardAddress struct {
	RewardAddress string `json:"rewardAddress"`
}

type MiningCandidate struct {
	Msg    string  `json:"msg"`
	B      float64 `json:"b"`
	Height uint64  `json:"h"`
	PK     string  `json:"pk"`
}
