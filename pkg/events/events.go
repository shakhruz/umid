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

package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"gitlab.com/umitop/umid/pkg/umi"
)

type iBlockchain interface {
	Subscribe(chan umi.Block)
}

type iMempool interface {
	Subscribe(chan *umi.Transaction)
}

type Events struct {
	sync.RWMutex
	blocks        chan umi.Block
	transactions  chan *umi.Transaction
	subscriptions map[umi.Address]map[chan<- []byte]struct{}
}

func NewEvents() *Events {
	return &Events{
		blocks:        make(chan umi.Block),
		transactions:  make(chan *umi.Transaction, 1_000),
		subscriptions: make(map[umi.Address]map[chan<- []byte]struct{}),
	}
}

func (events *Events) SubscribeTo(blockchain iBlockchain) {
	blockchain.Subscribe(events.blocks)
}

func (events *Events) SubscribeTo2(mempool iMempool) {
	mempool.Subscribe(events.transactions)
}

func (events *Events) Subscribe(address umi.Address, ch chan<- []byte) {
	events.Lock()
	defer events.Unlock()

	subscription, ok := events.subscriptions[address]
	if !ok {
		subscription = make(map[chan<- []byte]struct{})
		events.subscriptions[address] = subscription
	}

	subscription[ch] = struct{}{}
}

func (events *Events) Unsubscribe(address umi.Address, ch chan<- []byte) {
	events.Lock()
	defer events.Unlock()

	if subscription, ok := events.subscriptions[address]; ok {
		delete(subscription, ch)

		if len(subscription) == 0 {
			delete(events.subscriptions[address], ch)
		}
	}
}

func (events *Events) Worker(ctx context.Context) {
	for {
		select {
		case block := <-events.blocks:
			events.processBlock(block)

		case transaction := <-events.transactions:
			events.processTransaction(transaction)

		case <-ctx.Done():
			return
		}
	}
}

func (events *Events) processBlock(block umi.Block) {
	for i, j := 0, block.TransactionCount(); i < j; i++ {
		transaction := block.Transaction(i)
		data, _ := json.MarshalIndent(transaction, "data: ", "  ")
		buf := new(bytes.Buffer)
		_, _ = fmt.Fprintf(buf, "event: transaction\ndata: %s\n\n", data)
		data = buf.Bytes()

		events.notify(transaction.Sender(), data)

		if transaction.HasRecipient() {
			events.notify(transaction.Recipient(), data)
		}

		if transaction.HasFee() {
			events.notify(transaction.FeeAddress(), data)
		}
	}
}

func (events *Events) processTransaction(transaction *umi.Transaction) {
	data, _ := json.MarshalIndent(transaction, "data: ", "  ")
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, "event: mempool\ndata: %s\n\n", data)
	data = buf.Bytes()

	events.notify(transaction.Sender(), data)

	if transaction.HasRecipient() {
		events.notify(transaction.Recipient(), data)
	}
}

func (events *Events) notify(address umi.Address, data []byte) {
	events.RLock()
	defer events.RUnlock()

	if subscription, ok := events.subscriptions[address]; ok {
		for ch := range subscription {
			select {
			case ch <- data:
			default:
			}
		}
	}
}
