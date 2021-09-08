// Copyright (c) 2021 UMI
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

package storage_test

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"gitlab.com/umitop/umid/pkg/ledger"
	. "gitlab.com/umitop/umid/pkg/storage"
	"gitlab.com/umitop/umid/pkg/umi"
)

type BlockchainMock struct {
	subscriptions []chan umi.Block
}

func NewBlockchainMock() *BlockchainMock {
	return &BlockchainMock{
		subscriptions: make([]chan umi.Block, 0, 1),
	}
}

func (bc *BlockchainMock) Subscribe(ch chan umi.Block) {
	bc.subscriptions = append(bc.subscriptions, ch)
}

func (bc *BlockchainMock) Notify(block umi.Block) {
	for _, ch := range bc.subscriptions {
		ch <- block
	}
}

type mockLedger struct {
	Transactions map[umi.Hash]struct{}
}

func NewLedgerMock() *mockLedger {
	return &mockLedger{
		Transactions: make(map[umi.Hash]struct{}),
	}
}

func (mock *mockLedger) Account(address umi.Address) (account *ledger.Account, ok bool) {
	return &ledger.Account{
		Type:             umi.Umi,
		Balance:          42,
		UpdatedAt:        uint32(time.Now().Unix()),
		InterestRate:     0,
		TransactionCount: 0,
	}, true
}

func (mock *mockLedger) HasTransaction(hash umi.Hash) bool {
	_, ok := mock.Transactions[hash]

	return ok
}

func TestMempool_Subscribe(t *testing.T) {
	t.Parallel()

	sender := umi.Address{}
	recipient := umi.Address{}

	_, _ = rand.Read(sender[:])
	_, _ = rand.Read(recipient[:])

	transaction := umi.NewTransaction().SetVersion(umi.TxV1Send).SetSender(sender).SetRecipient(recipient).SetAmount(42)
	mempool := NewMempool()
	mempool.SetLedger(&mockLedger{})

	ch := make(chan *umi.Transaction, 1)

	mempool.Subscribe(ch)

	if err := mempool.Push(transaction); err != nil {
		t.Fatal(err)
	}

	transaction.SetAmount(42)
	if err := mempool.Push(transaction); err == nil {
		t.Errorf("must return error")
	}

	if len(ch) != 1 {
		t.Errorf("len must be 1, got %d", len(ch))
	}
}

func TestMempool_Push(t *testing.T) {
	t.Parallel()

	sender := umi.Address{}
	recipient := umi.Address{}

	_, _ = rand.Read(sender[:])
	_, _ = rand.Read(recipient[:])

	transaction := umi.NewTransaction()
	transaction.SetVersion(umi.TxV1Send).SetSender(sender).SetRecipient(recipient).SetAmount(42)

	mempool := NewMempool()
	mempool.SetLedger(&mockLedger{})

	if err := mempool.Push(transaction); err != nil {
		t.Fatal(err)
	}

	if err := mempool.Push(transaction); err == nil {
		t.Fatal("must return error")
	}
}

func TestMempool_Mempool(t *testing.T) {
	t.Parallel()

	sender := umi.Address{}
	recipient := umi.Address{}

	_, _ = rand.Read(sender[:])
	_, _ = rand.Read(recipient[:])

	transaction := umi.NewTransaction()
	transaction.SetVersion(umi.TxV1Send).SetSender(sender).SetRecipient(recipient).SetAmount(42)

	block := umi.NewBlock().SetVersion(1).SetTransactionCount(1)
	block = append(block, transaction...)
	block = append(block, make([]byte, 118)...)

	blockchainMock := NewBlockchainMock()
	ledgerMock := NewLedgerMock()

	mempool := NewMempool()
	mempool.SetLedger(ledgerMock)
	mempool.SubscribeTo(blockchainMock)

	_ = mempool.Push(transaction)

	transactions := mempool.Mempool()
	if len(transactions) != 1 {
		t.Errorf("totalCount must be 1, got %d", len(transactions))
	}

	blockchainMock.Notify(block)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	mempool.Worker(ctx)
	cancel()

	transactions = mempool.Mempool()

	if len(transactions) != 0 {
		t.Errorf("totalCount must be 0, got %d", len(transactions))
	}

	ledgerMock.Transactions[transaction.Hash()] = struct{}{}

	if err := mempool.Push(transaction); err == nil {
		t.Fatal("must return error")
	}

	mempool.ParseBlock(block)
}

func TestMempool_UnconfirmedBalance(t *testing.T) {
	t.Parallel()

	amount := uint64(42)
	sender := umi.Address{}
	recipient := umi.Address{}

	_, _ = rand.Read(sender[:])
	_, _ = rand.Read(recipient[:])

	transaction := umi.NewTransaction()
	transaction.SetVersion(umi.TxV1Send).SetSender(sender).SetRecipient(recipient).SetAmount(amount)

	mempool := NewMempool()
	mempool.SetLedger(&mockLedger{})

	if err := mempool.Push(transaction); err != nil {
		t.Fatal(err)
	}

	senderBalance := mempool.UnconfirmedBalance(sender)
	if senderBalance != -int64(amount) {
		t.Errorf("expected %d, got %d", -int64(amount), senderBalance)
	}

	recipientBalance := mempool.UnconfirmedBalance(recipient)
	if recipientBalance != int64(amount) {
		t.Errorf("expected %d, got %d", int64(amount), recipientBalance)
	}

	noneBalance := mempool.UnconfirmedBalance(umi.Address{})
	if noneBalance != int64(0) {
		t.Errorf("expected %d, got %d", int64(0), noneBalance)
	}
}

func TestMempool_Transactions(t *testing.T) {
	t.Parallel()

	sender := umi.Address{}
	recipient := umi.Address{}

	_, _ = rand.Read(sender[:])
	_, _ = rand.Read(recipient[:])

	transaction := umi.NewTransaction()
	transaction.SetVersion(umi.TxV1Send).SetSender(sender).SetRecipient(recipient).SetAmount(42)

	mempool := NewMempool()
	mempool.SetLedger(&mockLedger{})

	if err := mempool.Push(transaction); err != nil {
		t.Fatal(err)
	}

	transactions := mempool.Transactions(sender)

	if len(transactions) != 1 {
		t.Errorf("expected %d, got %d", 1, len(transactions))
	}

	transactions = mempool.Transactions(recipient)

	if len(transactions) != 1 {
		t.Errorf("expected %d, got %d", 1, len(transactions))
	}
}
