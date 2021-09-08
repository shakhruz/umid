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

package restapi

import (
	"gitlab.com/umitop/umid/pkg/events"
	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/storage"
)

type RestAPI struct {
	blockchain storage.IBlockchain
	ledger     *ledger.Ledger
	mempool    *storage.Mempool
	index      *storage.Index
	events     *events.Events
}

func NewRestAPI() *RestAPI {
	return &RestAPI{}
}

func (restApi *RestAPI) SetBlockchain(blockchain storage.IBlockchain) {
	restApi.blockchain = blockchain
}

func (restApi *RestAPI) SetLedger(ledger1 *ledger.Ledger) {
	restApi.ledger = ledger1
}

func (restApi *RestAPI) SetMempool(mempool *storage.Mempool) {
	restApi.mempool = mempool
}

func (restApi *RestAPI) SetIndex(index *storage.Index) {
	restApi.index = index
}

func (restApi *RestAPI) SetEvents(event *events.Events) {
	restApi.events = event
}
