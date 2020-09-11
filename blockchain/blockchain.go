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
	"encoding/hex"
	"errors"
	"sync"
	"umid/umid"
)

const txQueueLen = 100_000

var errTooManyRequests = errors.New("too many requests")

// Blockchain ...
type Blockchain struct {
	storage      umid.IStorage
	transaction  chan []byte
	approvedKeys map[string]struct{}
}

// NewBlockchain ...
func NewBlockchain() *Blockchain {
	keys := []string{
		"45885c9687d799a4d1f4d786d8639274d293ed024ad7f7a436715d6217c9f72b",
	}

	bc := &Blockchain{
		transaction:  make(chan []byte, txQueueLen),
		approvedKeys: make(map[string]struct{}, len(keys)),
	}

	for _, key := range keys {
		b, _ := hex.DecodeString(key)
		bc.approvedKeys[string(b)] = struct{}{}
	}

	return bc
}

// SetStorage ...
func (bc *Blockchain) SetStorage(db umid.IStorage) *Blockchain {
	bc.storage = db

	return bc
}

// Worker ...
func (bc *Blockchain) Worker(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
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
