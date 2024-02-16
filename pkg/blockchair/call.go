package blockchair

import (
	"fmt"
)

func (c *Client) GetAddressBTC(address string) (*RawAddress, error) {
	obj := new(AddressResponse)
	err := c.do("GET", "/bitcoin/dashboards/address/"+address, nil, obj)
	if err != nil {
		return nil, err
	} else if err := parseContext(obj.Context); err != nil {
		return nil, err
	}

	res, ok := obj.Data[address]
	if !ok {
		return nil, fmt.Errorf("unable to find address %s", address)
	} else if res.Address == nil {
		return nil, fmt.Errorf("nil address for %s", address)
	}

	return res.Address, nil
}

func (c *Client) GetTxBTC(txid string) (*TxInfo, error) {
	obj := new(TxResponse)
	err := c.do("GET", "/bitcoin/dashboards/transaction/"+txid, nil, obj)
	if err != nil {
		return nil, err
	} else if err := parseContext(obj.Context); err != nil {
		return nil, err
	}

	res, ok := obj.Data[txid]
	if !ok {
		return nil, fmt.Errorf("unable to find tx %s", txid)
	}

	return res, nil
}

func (c *Client) BroadcastTxBTC(data string) (string, error) {
	body := map[string]string{"data": data, "key": c.key}
	obj := new(BroadcastResponse)
	err := c.do("POST", "/bitcoin/push/transaction", body, obj)
	if err != nil {
		return "", err
	} else if err := parseContext(obj.Context); err != nil {
		return "", err
	} else if obj.Data == nil {
		return "", fmt.Errorf("nil data response for %v", obj)
	} else if len(obj.Data.TxID) == 0 {
		return "", fmt.Errorf("empty txid response for %v: %v", obj, obj.Data)
	}

	return obj.Data.TxID, nil
}
