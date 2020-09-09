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
	"umid/umid"

	"github.com/gorilla/websocket"
)

const (
	workerQueueLen = 1024
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
	Result  interface{}     `json:"result,omitempty"`
	Error   *respError      `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"`
}

type respError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RPC ...
type RPC struct {
	ctx           context.Context
	wg            *sync.WaitGroup
	blockchain    umid.IBlockchain
	upgrader      websocket.Upgrader
	queue         chan rawRequest
	methods       map[string]func(umid.IBlockchain, json.RawMessage, *response)
	notifications map[string]func(umid.IBlockchain, json.RawMessage)
}

// NewRPC ...
func NewRPC(ctx context.Context, wg *sync.WaitGroup, bc umid.IBlockchain) *RPC {
	return &RPC{
		ctx:           ctx,
		wg:            wg,
		blockchain:    bc,
		upgrader:      websocket.Upgrader{},
		queue:         make(chan rawRequest, workerQueueLen),
		methods:       make(map[string]func(umid.IBlockchain, json.RawMessage, *response)),
		notifications: make(map[string]func(umid.IBlockchain, json.RawMessage)),
	}
}

func newResponse(req request) *response {
	return &response{
		JSONRPC: "2.0",
		Result:  nil,
		Error:   nil,
		ID:      req.ID,
	}
}
