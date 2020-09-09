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
	"umid/umid"
)

const txQueueLen = 100_000

var errTooManyRequests = errors.New("too many requests")

// Blockchain ...
type Blockchain struct {
	ctx          context.Context
	wg           *sync.WaitGroup
	storage      umid.IStorage
	transaction  chan []byte
	approvedKeys map[string]struct{}
}

// NewBlockchain ...
func NewBlockchain(ctx context.Context, wg *sync.WaitGroup, db umid.IStorage) *Blockchain {
	return &Blockchain{
		ctx:         ctx,
		wg:          wg,
		storage:     db,
		transaction: make(chan []byte, txQueueLen),
		approvedKeys: map[string]struct{}{
			"\x45\x88\x5c\x96\x87\xd7\x99\xa4\xd1\xf4\xd7\x86\xd8\x63\x92\x74\xd2\x93\xed\x02\x4a\xd7\xf7\xa4\x36\x71\x5d\x62\x17\xc9\xf7\x2b": {},
		},
	}
}

// Worker ...
func (bc *Blockchain) Worker() {
	bc.wg.Add(1)
	defer bc.wg.Done()

	for {
		select {
		case <-bc.ctx.Done():
			return
		case t := <-bc.transaction:
			_ = bc.storage.AddTransaction(t)
		}
	}
}

// Mempool ...
func (bc *Blockchain) Mempool() (umid.IMempool, error) {
	return bc.storage.Mempool()
}
