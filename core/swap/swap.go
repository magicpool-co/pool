package swap

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/types"
)

var (
	ErrSwapNotReady = fmt.Errorf("swap not ready")
	ErrSwapExpired  = fmt.Errorf("swap expired")
	ErrSwapDust     = fmt.Errorf("swap is dust")
)

type Client struct {
	shuttleflow *ShuttleflowClient
	fusion      *FusionClient
}

func New() *Client {
	client := &Client{
		shuttleflow: NewShuttleflowClient(),
		fusion:      NewFusionClient(),
	}

	return client
}

// eSpace

func (c *Client) InitiateCFXToESpaceSwap(cfxNode, ethNode types.PayoutNode, amount *big.Int) (string, error) {
	return "", nil
}

func (c *Client) FinalizeCFXToESpaceSwap(ethNode types.PayoutNode) (string, error) {
	return "", nil
}

// shuttleflow

func (c *Client) InitiateCFXToBSCSwap(cfxNode, bscNode types.PayoutNode, amount *big.Int) (string, error) {
	return c.shuttleflow.InitiateSwapFromCFX(cfxNode, bscNode.Address(), amount)
}

func (c *Client) FinalizeCFXToBSCSwap(bscNode types.PayoutNode, txid string) (string, error) {
	return c.shuttleflow.FinalizeSwapFromCFX(bscNode, txid)
}

func (c *Client) InitiateBSCToCFXSwap(cfxNode, bscNode types.PayoutNode, amount *big.Int) (string, error) {
	return c.shuttleflow.InitiateSwapFromBSC(bscNode, cfxNode.Address(), amount)
}

func (c *Client) FinalizeBSCToCFXSwap(cfxNode types.PayoutNode, txid string) (string, error) {
	return c.shuttleflow.FinalizeSwapFromBSC(cfxNode, txid)
}

// fusion

func (c *Client) InitiateFLUXToBSCSwap(fluxNode, bscNode types.PayoutNode, inputs []*types.TxInput) (string, error) {
	return c.fusion.InitiateSwapFromFlux(fluxNode, bscNode.Address(), inputs)
}

func (c *Client) InitiateBSCToFLUXSwap(fluxNode, bscNode types.PayoutNode, amount *big.Int) (string, error) {
	return c.fusion.InitiateSwapFromBSC(bscNode, fluxNode.Address(), amount)
}

func (c *Client) FinalizeFLUXSwap(txid string) error {
	return c.fusion.FinalizeSwap(txid)
}
