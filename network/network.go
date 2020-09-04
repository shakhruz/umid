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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"umid/umid"
)

const (
	pullIntervalSec = 5
	pushIntervalSec = 5
)

// Network ...
type Network struct {
	ctx        context.Context
	wg         *sync.WaitGroup
	blockchain umid.IBlockchain
	client     *http.Client
}

// NewNetwork ...
func NewNetwork(ctx context.Context, wg *sync.WaitGroup, bc umid.IBlockchain) *Network {
	return &Network{
		ctx:        ctx,
		wg:         wg,
		blockchain: bc,
		client:     newClient(),
	}
}

// Worker ...
func (net *Network) Worker() {
	net.wg.Add(1)
	defer net.wg.Done()

	push := time.NewTicker(pushIntervalSec * time.Second)
	defer push.Stop()

	go net.puller()

	for {
		select {
		case <-net.ctx.Done():
			return
		case <-push.C:
			break
		}
	}
}

func (net *Network) puller() {
	net.wg.Add(1)
	defer net.wg.Done()

	for {
		select {
		case <-net.ctx.Done():
			return
		default:
			break
		}

		if net.pull() {
			continue
		}

		time.Sleep(pullIntervalSec * time.Second)
	}
}

func (net *Network) pull() (ok bool) {
	lstBlkHeight, err := net.blockchain.LastBlockHeight()
	if err != nil {
		return
	}

	const tpl = `{"jsonrpc":"2.0","method":"listBlocks","params":{"height":%d},"id":"%d"}`
	jsn := fmt.Sprintf(tpl, lstBlkHeight+1, time.Now().UnixNano())

	host := "https://mainnet.umi.top"
	if val, ok := os.LookupEnv("NETWORK"); ok && val == "testnet" {
		host = "https://testnet.umi.top"
	}

	if val, ok := os.LookupEnv("PEER"); ok {
		host = val
	}

	req, _ := http.NewRequestWithContext(net.ctx, "POST", fmt.Sprintf("%s/json-rpc", host), strings.NewReader(jsn))

	resp, err := net.client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err := resp.Body.Close(); err != nil {
		log.Println(err.Error())
	}

	if err != nil {
		log.Println(err.Error())

		return
	}

	rez := &struct{ Result []string }{}
	if err := json.Unmarshal(body, rez); err != nil {
		log.Println(err.Error())

		return
	}

	for _, s := range rez.Result {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			log.Println(err.Error())

			return
		}

		if err := net.blockchain.AddBlock(b); err != nil {
			log.Println(err.Error())

			return
		}
	}

	return len(rez.Result) > 0
}
