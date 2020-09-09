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

package network

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	pushIntervalSec = 5
	pushLimitTxs    = 10_000
)

func (net *Network) pusher() {
	net.wg.Add(1)
	defer net.wg.Done()

	for {
		select {
		case <-net.ctx.Done():
			return
		default:
			break
		}

		net.push()
		time.Sleep(pushIntervalSec * time.Second)
	}
}

func (net *Network) push() {
	mem, err := net.blockchain.Mempool()
	if err != nil {
		return
	}

	type params struct {
		Base64 []byte `json:"base64"`
	}

	type request struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  params `json:"params"`
		ID      int64  `json:"id"`
	}

	txs := make([]request, 0)

	for mem.Next() {
		if len(txs) > pushLimitTxs {
			break
		}

		txs = append(txs, request{
			JSONRPC: "2.0",
			Method:  "sendTransaction",
			Params:  params{mem.Value()},
			ID:      time.Now().UnixNano(),
		})
	}

	mem.Close()

	if len(txs) == 0 {
		return
	}

	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)

	jsn, _ := json.Marshal(txs)
	_, _ = gz.Write(jsn)
	_ = gz.Close()

	req, _ := http.NewRequestWithContext(net.ctx, "POST", peer(), buf)
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := net.client.Do(req)
	if err != nil {
		return
	}

	_, _ = io.Copy(ioutil.Discard, resp.Body)
	_ = resp.Body.Close()
}
