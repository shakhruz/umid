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
	"fmt"
	"net/http"
	"os"
	"sync"
	"umid/umid"
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
	if os.Getenv("PEER") != "none" {
		go net.puller()
		go net.pusher()
	}
}

func peer() (url string) {
	switch os.Getenv("NETWORK") {
	case "testnet":
		url = "https://testnet.umi.top"
	default:
		url = "https://mainnet.umi.top"
	}

	if val, ok := os.LookupEnv("PEER"); ok {
		url = val
	}

	return fmt.Sprintf("%s/json-rpc", url)
}
