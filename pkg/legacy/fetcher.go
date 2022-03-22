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
	"log"
	"net/http"
	"time"

	"gitlab.com/umitop/umid/pkg/config"
	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/nft"
	"gitlab.com/umitop/umid/pkg/umi"
)

type Fetcher struct {
	config     *config.Config
	client     *http.Client
	confirmer  *ledger.ConfirmerLegacy
	nftStorage *nft.Storage
}

func NewFetcher(conf *config.Config) *Fetcher {
	return &Fetcher{
		config: conf,
		client: newClient(),
	}
}

func (fetcher *Fetcher) SetConfirmer(confirmer *ledger.ConfirmerLegacy) {
	fetcher.confirmer = confirmer
}

func (fetcher *Fetcher) SetNftStorage(nftStorage *nft.Storage) {
	fetcher.nftStorage = nftStorage
}

func (fetcher *Fetcher) Worker(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for {
				if fetcher.fetchBlocks(ctx) < 1 {
					break
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func (fetcher *Fetcher) Worker2(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for {
				if fetcher.fetchNfts(ctx) < 1 {
					break
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func newRequestBody(height uint32, limit int) *bytes.Buffer {
	buffer := new(bytes.Buffer)
	requestBody := struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		ID      string `json:"id"`
		Params  struct {
			Height uint32 `json:"height"`
			Limit  int    `json:"limit"`
		} `json:"params"`
	}{
		JSONRPC: "2.0",
		Method:  "listBlocks",
		ID:      "1",
		Params: struct {
			Height uint32 `json:"height"`
			Limit  int    `json:"limit"`
		}{
			Height: height,
			Limit:  limit,
		},
	}

	_ = json.NewEncoder(buffer).Encode(requestBody)

	return buffer
}

func (fetcher *Fetcher) fetchNfts(ctx context.Context) int {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()

	height := fetcher.nftStorage.Count()
	url := fmt.Sprintf("%s/api/nfts?raw=true&height=%d", fetcher.config.Peer, height)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("fetch NFT error: %v", err)

		return -1
	}

	response, err := fetcher.client.Do(request)
	if err != nil {
		log.Printf("fetch NFT error: %v", err)

		return -1
	}

	defer response.Body.Close()

	responseBody := struct {
		Data  *[][]byte        `json:"data,omitempty"`
		Error *json.RawMessage `json:"error,omitempty"`
	}{}

	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		log.Printf("fetch NFT error: %v", err)

		return -1
	}

	if responseBody.Data == nil || len(*responseBody.Data) == 0 {
		return 0
	}

	if err := fetcher.nftStorage.AppendData((*responseBody.Data)[0]); err != nil {
		log.Printf("sync NFT error: %v", err)
	}

	return len(*responseBody.Data)
}

func (fetcher *Fetcher) fetchBlocks(ctx context.Context) int {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	url := fmt.Sprintf("%s/json-rpc", fetcher.config.Peer)
	height := fetcher.confirmer.BlockHeight + 1
	limit := 10_000
	requestBody := newRequestBody(height, limit)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		log.Printf("fetch error: %v", err)

		return -1
	}

	response, err := fetcher.client.Do(request)
	if err != nil {
		log.Printf("fetch error: %v", err)

		return -1
	}

	defer response.Body.Close()

	responseBody := struct {
		Result [][]byte `json:"result"`
	}{}

	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		log.Printf("fetch error: %v", err)

		return -1
	}

	if len(responseBody.Result) == 0 {
		return 0
	}

	// log.Printf("скачано %d блоков, начиная с %d", len(responseBody.Result), height)

	fetcher.parseBlocks(responseBody.Result)

	return len(responseBody.Result)
}

func (fetcher *Fetcher) parseBlocks(blocks [][]byte) {
	for _, block := range blocks {
		blk := (umi.BlockLegacy)(block)

		if err := blk.Verify(); err != nil {
			return
		}

		for i, j := 0, blk.TransactionCount(); i < j; i++ {
			if err := blk.Transaction(i).Verify(); err != nil {
				return
			}
		}

		err := fetcher.confirmer.AppendBlockLegacy(block)
		if err != nil {
			log.Printf("error: %v", err)

			return
		}
	}
}
