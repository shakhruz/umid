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
	"encoding/json"
	"log"
)

var (
	errParseError     = []byte(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}`)
	errInvalidRequest = []byte(`{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`)
	errInternalError  = []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"},"id":null}`)
	errInvalidParams  = &respError{Code: -32602, Message: "Invalid params"}
)

func (rpc *RPC) init() {
	rpc.methods["getBalance"] = rpc.getBalance
	rpc.methods["listStructures"] = rpc.listStructures
	rpc.methods["getStructure"] = rpc.getStructure
	rpc.methods["sendTransaction"] = rpc.sendTransaction
	rpc.methods["listTransactions"] = rpc.listTransactions
	rpc.methods["listBlocks"] = rpc.listBlocks
}

// Worker ...
func (rpc *RPC) Worker() {
	rpc.wg.Add(1)
	defer rpc.wg.Done()

	rpc.init()

	for {
		select {
		case <-rpc.ctx.Done():
			return
		case q := <-rpc.queue:
			select {
			case <-q.ctx.Done():
				continue
			default:
				break
			}

			if q.req == nil || len(q.req) == 0 {
				q.res <- errInvalidRequest

				continue
			}

			if string(q.req[0]) == "[" {
				q.res <- rpc.parseBatch(q.req)

				continue
			}

			q.res <- rpc.parseSingle(q.req)
		}
	}
}

func (rpc *RPC) parseSingle(data []byte) []byte {
	var req request

	if err := json.Unmarshal(data, &req); err != nil {
		return errParseError
	}

	if req.JSONRPC != "2.0" || req.Method == "" {
		return errInvalidRequest
	}

	if req.ID == nil {
		rpc.callNotification(req)

		return nil
	}

	return rpc.callMethod(req)
}

func (rpc *RPC) parseBatch(data []byte) (b []byte) {
	var batch []json.RawMessage

	if err := json.Unmarshal(data, &batch); err != nil {
		return errParseError
	}

	if len(batch) == 0 {
		return errInvalidRequest
	}

	res := make([]json.RawMessage, 0, len(batch))

	for _, raw := range batch {
		r := rpc.parseSingle(raw)
		if r != nil {
			res = append(res, r)
		}
	}

	if len(res) != 0 {
		var err error
		if b, err = json.Marshal(res); err != nil {
			b = errInternalError
		}
	}

	return b
}

func (rpc *RPC) callNotification(req request) {
	if notification, ok := rpc.notifications[req.Method]; ok {
		notification(req.Params)
	}
}

func (rpc *RPC) callMethod(req request) (b []byte) {
	method, ok := rpc.methods[req.Method]
	if !ok {
		method = func(_ json.RawMessage, res *response) {
			res.Error = &respError{
				Code:    -32601,
				Message: "Method not found",
			}
		}
	}

	res := newResponse(req)
	method(req.Params, res)

	var err error
	if b, err = json.Marshal(res); err != nil {
		log.Println(err.Error())

		b = errInternalError
	}

	return b
}
