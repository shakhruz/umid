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

package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/umi"
)

type iSubscriber interface {
	Subscribe(chan umi.Block)
}

type iLedger interface {
	Account(address umi.Address) (account *ledger.Account, ok bool)
	Structure(prefix umi.Prefix) (structure *ledger.Structure, ok bool)
	HasTransaction(hash umi.Hash) bool
}

type state struct {
	balance      int64
	transactions map[umi.Hash]*umi.Transaction
}

type Mempool struct {
	sync.RWMutex
	ledger        iLedger
	blocks        chan umi.Block
	addresses     map[umi.Address]*state
	transactions  map[umi.Hash]*umi.Transaction
	subscriptions []chan *umi.Transaction
}

var ErrMempool = errors.New("mempool")

func NewMempool() *Mempool {
	return &Mempool{
		blocks:        make(chan umi.Block, 64),
		addresses:     make(map[umi.Address]*state),
		transactions:  make(map[umi.Hash]*umi.Transaction),
		subscriptions: make([]chan *umi.Transaction, 0, 2),
	}
}

func (mempool *Mempool) SetLedger(ledger1 iLedger) {
	mempool.ledger = ledger1
}

func (mempool *Mempool) Subscribe(ch chan *umi.Transaction) {
	mempool.subscriptions = append(mempool.subscriptions, ch)
}

func (mempool *Mempool) SubscribeTo(subscriber iSubscriber) {
	subscriber.Subscribe(mempool.blocks)
}

func (mempool *Mempool) Worker(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case block := <-mempool.blocks:
			mempool.ParseBlock(block)

		case <-ticker.C:
			mempool.cleanup()

		case <-ctx.Done():
			return
		}
	}
}

func (mempool *Mempool) Push(transaction umi.Transaction) error {
	hash := transaction.Hash()

	mempool.Lock()
	defer mempool.Unlock()

	if _, ok := mempool.transactions[hash]; ok {
		return fmt.Errorf("%w: tranasction in mempool", ErrMempool)
	}

	if mempool.ledger.HasTransaction(hash) {
		return fmt.Errorf("%w: tranasction confirmed", ErrMempool)
	}

	senderAccount, ok := mempool.ledger.Account(transaction.Sender())
	if !ok {
		return fmt.Errorf("%w: sender account not found", ErrMempool)
	}

	if senderAccount.BalanceAt(uint32(time.Now().Unix())) < transaction.Amount() {
		return fmt.Errorf("%w: insufficient funds", ErrMempool)
	}

	mempool.insert(hash, &transaction)

	return nil
}

func (mempool *Mempool) UnconfirmedBalance(address umi.Address) int64 {
	mempool.RLock()
	defer mempool.RUnlock()

	if addressState, ok := mempool.addresses[address]; ok {
		return addressState.balance
	}

	return 0
}

func (mempool *Mempool) Transactions(address umi.Address) (transactions []*umi.Transaction) {
	mempool.RLock()
	defer mempool.RUnlock()

	transactions = make([]*umi.Transaction, 0)

	if addressState, ok := mempool.addresses[address]; ok {
		for _, transaction := range addressState.transactions {
			transactions = append(transactions, transaction)
		}
	}

	return transactions
}

func (mempool *Mempool) Mempool() (transactions []*umi.Transaction) {
	mempool.RLock()
	defer mempool.RUnlock()

	transactions = make([]*umi.Transaction, 0, len(mempool.transactions))

	for _, transaction := range mempool.transactions {
		transactions = append(transactions, transaction)
	}

	return transactions
}

func (mempool *Mempool) ParseBlock(block umi.Block) {
	mempool.Lock()
	defer mempool.Unlock()

	for i, txCount := 0, block.TransactionCount(); i < txCount; i++ {
		transaction := block.Transaction(i)
		hash := transaction.Hash()

		mempool.remove(hash)
	}
}

func (mempool *Mempool) insert(hash umi.Hash, transaction *umi.Transaction) {
	mempool.transactions[hash] = transaction

	amount := int64(transaction.Amount())
	sender := transaction.Sender()

	senderState, ok := mempool.addresses[sender]
	if !ok {
		senderState = &state{transactions: make(map[umi.Hash]*umi.Transaction)}
		mempool.addresses[sender] = senderState
	}

	senderState.balance -= amount
	senderState.transactions[hash] = transaction

	if transaction.HasRecipient() {
		recipient := transaction.Recipient()

		recipientState, ok := mempool.addresses[recipient]
		if !ok {
			recipientState = &state{transactions: make(map[umi.Hash]*umi.Transaction)}
			mempool.addresses[recipient] = recipientState
		}

		recipientState.balance += amount
		recipientState.transactions[hash] = transaction
	}

	mempool.notify(transaction)
}

func (mempool *Mempool) remove(hash umi.Hash) {
	transaction, ok := mempool.transactions[hash]
	if !ok {
		return
	}

	delete(mempool.transactions, hash)

	amount := int64(transaction.Amount())
	sender := transaction.Sender()

	senderState := mempool.addresses[sender]
	senderState.balance += amount

	delete(senderState.transactions, hash)

	if len(senderState.transactions) == 0 {
		delete(mempool.addresses, sender)
	}

	if transaction.HasRecipient() {
		recipient := transaction.Recipient()

		recipientState := mempool.addresses[recipient]
		recipientState.balance += amount

		delete(recipientState.transactions, hash)

		if len(recipientState.transactions) == 0 {
			delete(mempool.addresses, recipient)
		}
	}
}

func (mempool *Mempool) notify(transaction *umi.Transaction) {
	for _, ch := range mempool.subscriptions {
		select {
		case ch <- transaction:
		default:
		}
	}
}

func (mempool *Mempool) cleanup() {
	timestamp := uint32(time.Now().Unix())

	mempool.Lock()
	defer mempool.Unlock()

	for hash, transaction := range mempool.transactions {
		txTimestamp := transaction.Timestamp()

		// Транзакция из будущего.
		if txTimestamp > timestamp {
			mempool.remove(hash)

			continue
		}

		// Просроченная транзакция.
		if timestamp-txTimestamp > 3600 {
			mempool.remove(hash)

			continue
		}

		// Баланс отправителя не существует.
		account, ok := mempool.ledger.Account(transaction.Sender())
		if !ok {
			mempool.remove(hash)

			continue
		}

		// На балансе недостаточно монет.
		if account.BalanceAt(timestamp) < transaction.Amount() {
			mempool.remove(hash)

			continue
		}

		// Структура получателя не существует.
		if _, ok := mempool.ledger.Structure(transaction.Recipient().Prefix()); !ok {
			mempool.remove(hash)

			continue
		}
	}
}
