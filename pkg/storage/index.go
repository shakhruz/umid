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
	"sync"

	"gitlab.com/umitop/umid/pkg/umi"
)

type Index struct {
	sync.RWMutex
	blocks    chan umi.Block
	addresses map[umi.Address]*[]uint64
}

func NewIndex() *Index {
	return &Index{
		blocks:    make(chan umi.Block, 64),
		addresses: make(map[umi.Address]*[]uint64),
	}
}

func (index *Index) TransactionsByAddress(address umi.Address) (txs *[]uint64, ok bool) {
	index.RLock()
	txs, ok = index.addresses[address]
	index.RUnlock()

	return txs, ok
}

func (index *Index) SubscribeTo(subscriber iSubscriber) {
	subscriber.Subscribe(index.blocks)
}

func (index *Index) Worker(ctx context.Context) {
	for {
		select {
		case block := <-index.blocks:
			index.processBlock(block)

		case <-ctx.Done():
			return
		}
	}
}

func (index *Index) processBlock(block umi.Block) {
	index.Lock()
	defer index.Unlock()

	for i, j := 0, block.TransactionCount(); i < j; i++ {
		transaction := block.Transaction(i)
		sender := transaction.Sender()
		tx := uint64(transaction.BlockHeight())<<16 | uint64(transaction.BlockTransactionIndex())

		senderTxs, ok := index.addresses[sender]
		if !ok {
			senderTxs = new([]uint64)
			index.addresses[sender] = senderTxs
		}

		*senderTxs = append(*senderTxs, tx)

		if transaction.HasRecipient() {
			recipient := transaction.Recipient()

			recipientTxs, ok := index.addresses[recipient]
			if !ok {
				recipientTxs = new([]uint64)
				index.addresses[recipient] = recipientTxs
			}

			*recipientTxs = append(*recipientTxs, tx)
		}

		if transaction.HasFee() {
			fee := transaction.FeeAddress()

			feeTxs, ok := index.addresses[fee]
			if !ok {
				feeTxs = new([]uint64)
				index.addresses[fee] = feeTxs
			}

			*feeTxs = append(*feeTxs, tx)
		}
	}
}
