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

func (rpc *RPC) listTransactions(raw json.RawMessage, res *response) {
	prm := new(struct {
		Address string `json:"address"`
	})

	if err := json.Unmarshal(raw, prm); err != nil || prm.Address == "" {
		res.Error = errInvalidParams

		return
	}

	txs, err := rpc.blockchain.TransactionsByAddress(prm.Address)
	if err != nil {
		log.Println(err.Error())

		res.Error = &respError{
			Code:    -32603,
			Message: "Internal error",
		}

		return
	}

	res.Result = txs
}
