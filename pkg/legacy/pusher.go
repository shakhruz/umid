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

package legacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gitlab.com/umitop/umid/pkg/config"
	"gitlab.com/umitop/umid/pkg/umi"
)

type iMempool interface {
	Mempool() []*umi.Transaction
	Subscribe(ch chan *umi.Transaction)
}

type Pusher struct {
	config  *config.Config
	client  *http.Client
	mempool iMempool
	queue   chan *umi.Transaction
}

func NewPusher(conf *config.Config, mempool iMempool) *Pusher {
	queue := make(chan *umi.Transaction, 1_000)
	mempool.Subscribe(queue)

	return &Pusher{
		config:  conf,
		client:  newClient(),
		mempool: mempool,
		queue:   queue,
	}
}

func (pusher *Pusher) Worker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case tx := <-pusher.queue:
			pusher.push([]*umi.Transaction{tx})

		case <-ticker.C:
			txs := pusher.mempool.Mempool()
			if len(txs) > 0 {
				pusher.push(txs)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (pusher *Pusher) push(txs []*umi.Transaction) {
	url := fmt.Sprintf("%s/json-rpc", pusher.config.Peer)

	for _, transaction := range txs {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		requestBody := newPushRequest(*transaction)
		request, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBody)

		response, err := pusher.client.Do(request)
		if err != nil {
			cancel()
			log.Println(err.Error())

			return
		}

		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()

		cancel()
	}
}

func newPushRequest(transaction []byte) *bytes.Buffer {
	buffer := new(bytes.Buffer)
	requestBody := struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		ID      string `json:"id"`
		Params  struct {
			Base64 []byte `json:"base64"`
		} `json:"params"`
	}{
		JSONRPC: "2.0",
		Method:  "sendTransaction",
		ID:      "1",
		Params: struct {
			Base64 []byte `json:"base64"`
		}{
			Base64: transaction,
		},
	}

	_ = json.NewEncoder(buffer).Encode(requestBody)

	return buffer
}
