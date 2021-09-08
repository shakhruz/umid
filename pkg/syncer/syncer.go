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

package syncer

import (
	"io"
	"net/http"

	"gitlab.com/umitop/umid/pkg/umi"
)

type iMempool interface {
	Push(transaction umi.Transaction) error
	Mempool() (transactions []*umi.Transaction)
}

type Syncer struct {
	mempool iMempool
}

func NewSyncer() *Syncer {
	return &Syncer{}
}

func (syncer *Syncer) SetMempool(mempool iMempool) {
	syncer.mempool = mempool
}

func (syncer *Syncer) Router(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/sync/mempool" {
		switch r.Method {
		case http.MethodGet:
			syncer.mempoolz()(w, r)
		case http.MethodPost:
			syncer.batch()(w, r)
		}
	}
}

func (syncer *Syncer) batch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactions := make([]umi.Transaction, 0)

		for {
			transaction := make(umi.Transaction, umi.TxLength)

			if _, err := io.ReadFull(r.Body, transaction); err != nil {
				break
			}

			if err := transaction.Verify(); err != nil {
				continue
			}

			transactions = append(transactions, transaction)
		}

		for _, transaction := range transactions {
			_ = syncer.mempool.Push(transaction)
		}
	}
}

func (syncer *Syncer) mempoolz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactions := syncer.mempool.Mempool()

		for _, transaction := range transactions {
			if _, err := w.Write(*transaction); err != nil {
				return
			}
		}
	}
}
