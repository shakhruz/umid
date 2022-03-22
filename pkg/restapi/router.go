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
	"fmt"
	"net/http"
	"runtime/metrics"
	"strings"

	"gitlab.com/umitop/umid/pkg/restapi/handler"
)

//nolint:funlen,gocognit,gocyclo,revive,cyclop // ...
func (restApi *RestAPI) Router(w http.ResponseWriter, r *http.Request) {
	var handlerFunc http.HandlerFunc

	path := r.URL.Path

	switch {
	case path == "/api/address:create":
		switch r.Method {
		case http.MethodPost:
			handlerFunc = handler.CreateAddress()
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodPost)
		}

	case path == "/api/transaction:create":
		switch r.Method {
		case http.MethodPost:
			handlerFunc = handler.CreateTransaction()
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodPost)
		}

	case path == "/api/mempool":
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.ListMempool(restApi.mempool)
		case http.MethodPost:
			handlerFunc = handler.PushMempool(restApi.mempool, restApi.nftMempool)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet, http.MethodPost)
		}

	case strings.HasPrefix(path, "/api/addresses/") && strings.HasSuffix(path, "/account"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.GetAccount(restApi.ledger, restApi.mempool)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/api/addresses/") && strings.HasSuffix(path, "/mempool"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.ListMempoolByAddress(restApi.mempool)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/api/addresses/") && strings.HasSuffix(path, "/transactions"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.ListTransactionsByAddress(restApi.blockchain, restApi.index)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/api/addresses/") && strings.HasSuffix(path, "/nfts"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.ListNftsByAddress(restApi.ledger)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case path == "/api/structures":
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.ListStructures(restApi.ledger)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/api/structures/"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.GetStructure(restApi.ledger)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case path == "/api/blocks":
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.ListBlocks(restApi.blockchain)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/api/blocks/") && strings.HasSuffix(path, "/transactions"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.ListTransactionsByBlock(restApi.blockchain)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/api/blocks/"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.GetBlock(restApi.blockchain)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/events/addresses/"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.SubscribeAddress(restApi.events)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/api/nfts/") && strings.HasSuffix(path, "/meta"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.GetNftMeta(restApi.nftStorage)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case strings.HasPrefix(path, "/api/nfts/"):
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.GetNft(restApi.nftStorage)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	case path == "/api/nfts":
		switch r.Method {
		case http.MethodGet:
			handlerFunc = handler.ListNft(restApi.nftStorage)
		default:
			handlerFunc = handler.MethodNotAllowed(http.MethodGet)
		}

	default:
		handlerFunc = handler.NotFound()
	}

	handlerFunc(w, r)
}

func (*RestAPI) Healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprint(w, "OK")
}

func (*RestAPI) Status(w http.ResponseWriter, _ *http.Request) {
	// Create a sample for the metric.
	sample := make([]metrics.Sample, 4)
	sample[0].Name = "/memory/classes/total:bytes"
	sample[1].Name = "/memory/classes/heap/free:bytes"
	sample[2].Name = "/sched/goroutines:goroutines"
	sample[3].Name = "/gc/heap/objects:objects"

	// Sample the metric.
	metrics.Read(sample)

	_, _ = fmt.Fprintf(w, "All memory mapped by the Go runtime: %d\n", sample[0].Value.Uint64())
	_, _ = fmt.Fprintf(w, "Memory that is completely free :     %d\n", sample[1].Value.Uint64())
	_, _ = fmt.Fprintf(w, "Count of live goroutines: %d\n", sample[2].Value.Uint64())
	_, _ = fmt.Fprintf(w, "Number of objects: %d\n", sample[3].Value.Uint64())
}
