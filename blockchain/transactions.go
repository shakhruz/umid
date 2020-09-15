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
	"context"
	"errors"
	"sync"

	"github.com/umitop/libumi"
)

const txQueueLen = 100_000

var (
	errInvalidLength   = errors.New("invalid length")
	errInvalidValue    = errors.New("invalid value")
	errTooManyRequests = errors.New("too many requests")
)

type iTransaction interface {
	Mempool(ctx context.Context) (raws <-chan []byte, err error)
	AddTxToMempool(raw []byte) error
	ListTxsByAddressAfterKey(adr []byte, key []byte, lim uint16) (raws [][]byte, err error)
	ListTxsByAddressBeforeKey(adr []byte, key []byte, lim uint16) (raws [][]byte, err error)
}

// Worker ...
func (bc *transaction) Worker(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case tx := <-bc.transaction:
			_ = bc.db.AddTxToMempool(tx)
		}
	}
}

// Mempool ...
func (bc *transaction) Mempool(ctx context.Context) (raws <-chan []byte, err error) {
	return bc.db.Mempool(ctx)
}

// AddTransaction ...
func (bc *transaction) AddTxToMempool(raw []byte) error {
	if err := bc.VerifyTransaction(raw); err != nil {
		return err
	}

	select {
	case bc.transaction <- raw:
		break
	default:
		return errTooManyRequests
	}

	return nil
}

// TransactionsByAddress ...
func (bc *transaction) ListTxsByAddressAfterKey(adr []byte, key []byte, lim uint16) (raws [][]byte, err error) {
	return bc.db.ListTxsByAddressAfterKey(adr, key, lim)
}

func (bc *transaction) ListTxsByAddressBeforeKey(adr []byte, key []byte, lim uint16) (raws [][]byte, err error) {
	return bc.db.ListTxsByAddressBeforeKey(adr, key, lim)
}

// VerifyTransaction ...
func (bc *transaction) VerifyTransaction(raw []byte) error {
	tx := (libumi.Transaction)(raw)

	if len(tx) != libumi.TxLength {
		return errInvalidLength
	}

	if tx.Version() == libumi.Basic {
		if !verifyValue(tx.Value()) {
			return errInvalidValue
		}
	}

	return libumi.VerifyTransaction(raw)
}

func verifyValue(val uint64) bool {
	const (
		minValue     = 1
		maxSafeValue = 90_071_992_547_409_91
	)

	return val >= minValue && val <= maxSafeValue
}

type transaction struct {
	transaction chan []byte
	db          iTransaction
}

func newTransaction(db iTransaction) transaction {
	return transaction{
		transaction: make(chan []byte, txQueueLen),
		db:          db,
	}
}
