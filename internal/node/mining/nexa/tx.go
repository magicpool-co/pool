package nexa

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/crypto/tx/nexatx"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://explorer.nexa.org/tx/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.nexa.org/address/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	conn, err := net.Dial("tcp", "rostrum.nexa.org:20001")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	request, err := rpc.NewRequest("blockchain.address.get_balance", node.address)
	if err != nil {
		return nil, err
	}

	msg, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(append(msg, '\n'))
	if err != nil {
		return nil, err
	}

	timer := time.NewTimer(time.Second * 5)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-timer.C:
			return nil, fmt.Errorf("rostrum timeout")
		default:
			var msg *rpc.Message
			err := json.Unmarshal(scanner.Bytes(), &msg)
			if err != nil {
				return nil, err
			} else if len(msg.Result) == 0 {
				return nil, nil
			}

			var data map[string]uint64
			err = json.Unmarshal(msg.Result, &data)
			if err != nil {
				return nil, err
			}

			balance := data["confirmed"] + data["unconfirmed"]

			return new(big.Int).SetUint64(balance), nil
		}
	}

	return nil, fmt.Errorf("rostrum empty results")
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	tx, err := node.getRawTransaction(txid)
	if err != nil {
		return nil, err
	}

	var height uint64
	if tx.BlockHash != "" {
		block, err := node.getBlock(tx.BlockHash)
		if err != nil {
			return nil, err
		}
		height = block.Height
	}

	var confirmed bool
	if height > 0 && tx.Confirmations > 0 {
		confirmed = true
	}

	res := &types.TxResponse{
		Hash:        txid,
		BlockNumber: height,
		Confirmed:   confirmed,
	}

	return res, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	const feeRate = 5

	baseTx := nexatx.NewTransaction(0, 0, node.prefix)
	rawTx, err := nexatx.GenerateTx(node.privKey, baseTx, inputs, outputs, feeRate)
	if err != nil {
		return "", "", err
	}
	tx := hex.EncodeToString(rawTx)
	txid := nexatx.CalculateTxIdem(tx)

	return txid, tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}
