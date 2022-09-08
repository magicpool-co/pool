package aetx

import (
	"io"

	"github.com/aeternity/rlp-go"
)

type signedTransaction struct {
	Signatures  [][]byte
	Transaction *transaction
}

// EncodeRLP implements rlp.Encoder
func (tx *signedTransaction) EncodeRLP(w io.Writer) error {
	// RLP serialize the wrapped Transaction into a plain bytearray.
	wrappedTxRLPBytes, err := rlp.EncodeToBytes(tx.Transaction)
	if err != nil {
		return err
	}

	// RLP Serialize the signedTransaction
	rlpRawMsg, err := buildRLPMessage(
		ObjectTagSignedTransaction,
		rlpMessageVersion,
		tx.Signatures,
		wrappedTxRLPBytes,
	)
	if err != nil {
		return err
	}

	_, err = w.Write(rlpRawMsg)
	if err != nil {
		return err
	}

	return nil
}

type signedTransactionRLP struct {
	ObjectTag         uint
	RlpMessageVersion uint
	Signatures        [][]byte
	WrappedTx         []byte
}

func (stx *signedTransactionRLP) ReadRLP(s *rlp.Stream) (err error) {
	var blob []byte
	if blob, err = s.Raw(); err != nil {
		return
	}
	if err = rlp.DecodeBytes(blob, stx); err != nil {
		return
	}
	return
}

// DecodeRLP implements rlp.Decoder
func (tx *signedTransaction) DecodeRLP(s *rlp.Stream) error {
	stx := &signedTransactionRLP{}
	if err := stx.ReadRLP(s); err != nil {
		return err
	}

	wtx := &transaction{}
	if err := rlp.DecodeBytes(stx.WrappedTx, wtx); err != nil {
		return err
	}

	tx.Signatures = stx.Signatures
	tx.Transaction = wtx

	return nil
}
