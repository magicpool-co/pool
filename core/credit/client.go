package credit

type Client struct {
}

func New() *Client {
	client := &Client{}

	return client
}

func (c *Client) UnlockBlocks() error {
	return nil
}

func (c *Client) CreditRounds() error {
	return nil
}
