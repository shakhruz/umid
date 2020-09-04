// Copyright (c) 2020 UMI
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package blockchain

import (
	"encoding/hex"
	"errors"
	"umid/umid"

	"github.com/umitop/libumi"
)

var errTxInvalidValue = errors.New("transaction: invalid value")

// AddTransaction ...
func (bc *Blockchain) AddTransaction(b []byte) error {
	if err := bc.VerifyTransaction(b); err != nil {
		return err
	}

	select {
	case bc.transaction <- b:
		break
	default:
		return errTooManyRequests
	}

	return nil
}

// TransactionsByAddress ...
func (bc *Blockchain) TransactionsByAddress(s string) ([]*umid.Transaction, error) {
	adr, err := libumi.NewAddressFromBech32(s)
	if err != nil {
		return nil, err
	}

	raw, err := bc.storage.TransactionsByAddress(adr)
	if err != nil {
		return nil, err
	}

	txs := make([]*umid.Transaction, 0, len(raw))

	for _, tx := range raw {
		t := &umid.Transaction{
			Hash:        hex.EncodeToString(tx.Hash),
			Height:      tx.Height,
			ConfirmedAt: tx.ConfirmedAt.Unix(),
			BlockHeight: tx.BlockHeight,
			BlockTxIdx:  tx.BlockTxIdx,
			Version:     tx.Version,
			Sender:      (*libumi.Address)(&tx.Sender).Bech32(),
			Value:       tx.Value,
			FeeValue:    tx.FeeValue,
			Structure:   tx.Structure,
		}

		if tx.Recipient != nil {
			t.Recipient = (*libumi.Address)(&tx.Recipient).Bech32()
		}

		if tx.FeeAddress != nil {
			t.FeeAddress = (*libumi.Address)(&tx.FeeAddress).Bech32()
		}

		txs = append(txs, t)
	}

	return txs, nil
}

// VerifyTransaction ...
func (bc *Blockchain) VerifyTransaction(b []byte) error {
	t := (libumi.Transaction)(b)

	if t.Value() == 0 {
		return errTxInvalidValue
	}

	return t.Verify()
}
