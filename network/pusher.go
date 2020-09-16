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
)

const (
	pushIntervalSec = 5
	pushTxsLimit    = 1000
)

func (net *Network) pusher(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pushIntervalSec * time.Second):
			legacyPush(ctx, net.client, net.blockchain)
		}
	}
}

func legacyPush(ctx context.Context, client *http.Client, bc iBlockchain) {
	txs := fetchMempool(bc)
	if len(txs) == 0 {
		return
	}

	for _, body := range txs {
		req, _ := http.NewRequestWithContext(ctx, "POST", peer(), bytes.NewReader(body))

		resp, err := client.Do(req)
		if err != nil {
			return
		}

		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

//nolint:deadcode,unused
func push(ctx context.Context, client *http.Client, bc iBlockchain) {
	txs := fetchMempool(bc)
	if len(txs) == 0 {
		return
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", peer(), marshalAndGzip(txs))
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	_, _ = io.Copy(ioutil.Discard, resp.Body)
	_ = resp.Body.Close()
}

func fetchMempool(bc iBlockchain) []json.RawMessage {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mem, err := bc.Mempool(ctx)
	if err != nil {
		return nil
	}

	txs := make([]json.RawMessage, 0)

	for tx := range mem {
		txs = append(txs, marshalPushRequest(tx))
		if len(txs) >= pushTxsLimit {
			break
		}
	}

	return txs
}

//nolint:unused
func marshalAndGzip(txs []json.RawMessage) *bytes.Buffer {
	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)
	jsn, _ := json.Marshal(txs)
	_, _ = gz.Write(jsn)
	_ = gz.Close()

	return buf
}

func marshalPushRequest(tx []byte) []byte {
	var b bytes.Buffer

	b.WriteString(`{"jsonrpc":"2.0","method":"sendTransaction","params":{"base64":"`)
	b.WriteString(base64.StdEncoding.EncodeToString(tx))
	b.WriteString(`"},"id":"1"}`)

	return b.Bytes()
}
