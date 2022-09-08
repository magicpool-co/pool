package aetx

import (
	"io"
	"math/big"

	"github.com/aeternity/rlp-go"
)

// Transaction represents a simple transaction where one party sends another AE
type transaction struct {
	SenderID    string
	RecipientID string
	Amount      *big.Int
	Fee         *big.Int
	Payload     []byte
	TTL         uint64
	Nonce       uint64
}

// EncodeRLP implements rlp.Encoder
func (tx *transaction) EncodeRLP(w io.Writer) error {
	// build id for the sender
	sID, err := buildIDTag(IDTagAccount, tx.SenderID)
	if err != nil {
		return err
	}

	// build id for the recipient
	rID, err := buildIDTag(IDTagAccount, tx.RecipientID)
	if err != nil {
		return err
	}

	// create the transaction
	rlpRawMsg, err := buildRLPMessage(
		ObjectTagSpendTransaction,
		rlpMessageVersion,
		sID,
		rID,
		tx.Amount,
		tx.Fee,
		tx.TTL,
		tx.Nonce,
		[]byte(tx.Payload))
	if err != nil {
		return err
	}

	_, err = w.Write(rlpRawMsg)
	if err != nil {
		return err
	}

	return nil
}

type spendRLP struct {
	ObjectTagSpendTransaction uint
	RlpMessageVersion         uint
	SenderID                  []uint8
	ReceiverID                []uint8
	Amount                    *big.Int
	Fee                       *big.Int
	TTL                       uint64
	Nonce                     uint64
	Payload                   []byte
}

func (stx *spendRLP) ReadRLP(s *rlp.Stream) (string, string, error) {
	blob, err := s.Raw()
	if err != nil {
		return "", "", err
	} else if err := rlp.DecodeBytes(blob, stx); err != nil {
		return "", "", err
	}

	sid, err := readIDTag(stx.SenderID)
	if err != nil {
		return "", "", err
	}

	rid, err := readIDTag(stx.ReceiverID)
	if err != nil {
		return "", "", err
	}

	return sid, rid, nil
}

// DecodeRLP implements rlp.Decoder
func (tx *transaction) DecodeRLP(s *rlp.Stream) error {
	stx := &spendRLP{}
	sID, rID, err := stx.ReadRLP(s)
	if err != nil {
		return err
	}

	tx.SenderID = sID
	tx.RecipientID = rID
	tx.Amount = stx.Amount
	tx.Fee = stx.Fee
	tx.TTL = stx.TTL
	tx.Nonce = stx.Nonce
	tx.Payload = stx.Payload

	return nil
}
