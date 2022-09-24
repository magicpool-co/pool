package payout

import (
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Client struct {
	pooldb *dbcl.Client
	nodes  map[string]types.PayoutNode
}

func New(pooldbClient *dbcl.Client, nodes map[string]types.PayoutNode) (*Client, error) {
	client := &Client{
		pooldb: pooldbClient,
		nodes:  nodes,
	}

	return client, nil
}

func (c *Client) InitiatePayouts() error {
	return nil
}

func (c *Client) FinalizePayouts() error {
	return nil
}
