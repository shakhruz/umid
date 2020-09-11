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
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"umid/umid"
)

const (
	pushIntervalSec = 5
	pushLimitTxs    = 10_000
)

func (net *Network) pusher(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pushIntervalSec * time.Second):
			push(ctx, net.client, net.blockchain)
		}
	}
}

func push(ctx context.Context, client *http.Client, bc umid.IBlockchain) {
	txs := prepareRequest(bc)
	if len(txs) == 0 {
		return
	}

	jsn, _ := json.Marshal(txs)

	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)
	_, _ = gz.Write(jsn)
	_ = gz.Close()

	req, _ := http.NewRequestWithContext(ctx, "POST", peer(), buf)
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	_, _ = io.Copy(ioutil.Discard, resp.Body)
	_ = resp.Body.Close()
}

func prepareRequest(bc umid.IBlockchain) []json.RawMessage {
	mem, err := bc.Mempool()
	if err != nil {
		return nil
	}

	defer mem.Close()

	buf := new(bytes.Buffer)
	txs := make([]json.RawMessage, 0)

	for mem.Next() {
		if len(txs) > pushLimitTxs {
			break
		}

		buf.Reset()
		buf.WriteString(`{"jsonrpc":"2.0","method":"sendTransaction","params":{"base64":"`)
		buf.WriteString(base64.StdEncoding.EncodeToString(mem.Value()))
		buf.WriteString(`"},"id":1}`)

		txs = append(txs, buf.Bytes())
	}

	return txs
}
