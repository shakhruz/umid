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

package jsonrpc

import (
	"context"
	"encoding/json"
	"sync"
	"umid/jsonrpc/method"
	"umid/umid"

	"github.com/gorilla/websocket"
)

const (
	workerQueueLen = 1024
)

var (
	errParseError     = []byte(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}`)
	errInvalidRequest = []byte(`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`)
	errInternalError  = []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"},"id":null}`)
)

type rawRequest struct {
	ctx context.Context
	req []byte
	res chan<- []byte
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"`
}

// Method ...
type Method func(bc umid.IBlockchain, params json.RawMessage) (result json.RawMessage, error json.RawMessage)

// RPC ...
type RPC struct {
	blockchain    umid.IBlockchain
	upgrader      websocket.Upgrader
	queue         chan rawRequest
	methods       map[string]Method
	notifications map[string]func(umid.IBlockchain, json.RawMessage)
}

// NewRPC ...
func NewRPC() *RPC {
	rpc := &RPC{
		upgrader:      websocket.Upgrader{},
		queue:         make(chan rawRequest, workerQueueLen),
		methods:       make(map[string]Method),
		notifications: make(map[string]func(umid.IBlockchain, json.RawMessage)),
	}

	rpc.methods["getBalance"] = method.GetBalance{}.Process
	rpc.methods["listStructures"] = method.ListStructures{}.Process
	rpc.methods["getStructure"] = method.GetStructure{}.Process
	rpc.methods["sendTransaction"] = method.SendTx{}.Process
	rpc.methods["listTransactions"] = method.ListTxs{}.Process
	rpc.methods["listBlocks"] = method.ListBlocks{}.Process

	return rpc
}

// SetBlockchain ...
func (rpc *RPC) SetBlockchain(bc umid.IBlockchain) *RPC {
	rpc.blockchain = bc

	return rpc
}

// Worker ...
func (rpc *RPC) Worker(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case q := <-rpc.queue:
			q.res <- processRequest(q.ctx, q.req, rpc)
		}
	}
}

func processRequest(ctx context.Context, req json.RawMessage, rpc *RPC) (res json.RawMessage) {
	select {
	case <-ctx.Done():
		return nil
	default:
		break
	}

	if len(req) == 0 {
		return errInvalidRequest
	}

	if string(req[0]) == "[" {
		return processBatch(req, rpc)
	}

	return processSingle(req, rpc)
}

func processSingle(data json.RawMessage, rpc *RPC) []byte {
	req := new(request)

	if err := json.Unmarshal(data, req); err != nil {
		return errParseError
	}

	if req.JSONRPC != "2.0" || req.Method == "" {
		return errInvalidRequest
	}

	if req.ID == nil {
		return nil
	}

	res, err := callMethod(req.Method, req.Params, rpc)

	return marshalResponse(res, err, req.ID)
}

func processBatch(req json.RawMessage, rpc *RPC) []byte {
	var requests []json.RawMessage

	if err := json.Unmarshal(req, &requests); err != nil {
		return errParseError
	}

	if len(requests) == 0 {
		return errInvalidRequest
	}

	responses := processBatchRequests(requests, rpc)

	if len(responses) == 0 {
		return nil
	}

	b, _ := json.Marshal(responses)

	return b
}

func processBatchRequests(requests []json.RawMessage, rpc *RPC) []json.RawMessage {
	res := make([]json.RawMessage, 0, len(requests))

	for _, request := range requests {
		if r := processSingle(request, rpc); r != nil {
			res = append(res, r)
		}
	}

	return res
}

func callMethod(name string, prm json.RawMessage, rpc *RPC) (result json.RawMessage, error json.RawMessage) {
	fn, ok := rpc.methods[name]
	if !ok {
		fn = func(_ umid.IBlockchain, _ json.RawMessage) (_ json.RawMessage, err json.RawMessage) {
			err, _ = json.Marshal(struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    -32601,
				Message: "Method not found",
			})

			return nil, err
		}
	}

	return fn(rpc.blockchain, prm)
}

func marshalResponse(result json.RawMessage, error json.RawMessage, id json.RawMessage) []byte {
	res := response{
		JSONRPC: "2.0",
		Result:  result,
		Error:   error,
		ID:      id,
	}

	b, err := json.Marshal(res)
	if err != nil {
		b = errInternalError
	}

	return b
}
