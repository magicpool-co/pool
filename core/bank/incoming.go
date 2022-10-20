package bank

import (
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) RegisterIncomingTx(node types.PayoutNode, txid string) (bool, error) {
	tx, err := node.GetTx(txid)
	if err != nil {
		return false, err
	} else if tx == nil || !tx.Confirmed {
		return false, nil
	}

	var utxos []*pooldb.UTXO
	switch node.GetAccountingType() {
	case types.AccountStructure:
		utxos = []*pooldb.UTXO{
			&pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    txid,
				Value:   dbcl.NullBigInt{Valid: true, BigInt: tx.Value},
				Active:  true,
			},
		}
	case types.UTXOStructure:
		utxos = make([]*pooldb.UTXO, 0)
		for _, output := range tx.Outputs {
			if output.Address != node.Address() {
				continue
			}

			utxo := &pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    txid,
				Index:   output.Index,
				Value:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(output.Value)},
				Active:  true,
			}
			utxos = append(utxos, utxo)
		}
	}

	err = pooldb.InsertUTXOs(c.pooldb.Writer(), utxos...)
	if err != nil {
		return true, err
	}

	return true, nil
}
