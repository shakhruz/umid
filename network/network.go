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
	"net/http"
	"os"
	"sync"
)

type iBlockchain interface {
	Mempool(ctx context.Context) (raws <-chan []byte, err error)
	GetLastBlockHeight() (height uint64, err error)
	GetLastBlockHash() (hash []byte, err error)
	AddBlock(raw []byte) error
}

// Network ...
type Network struct {
	blockchain iBlockchain
	client     *http.Client
}

// NewNetwork ...
func NewNetwork() *Network {
	return &Network{
		client: newClient(),
	}
}

// SetBlockchain ...
func (net *Network) SetBlockchain(bc iBlockchain) *Network {
	net.blockchain = bc

	return net
}

// Worker ...
func (net *Network) Worker(ctx context.Context, wg *sync.WaitGroup) {
	if os.Getenv("PEER") == "none" {
		return
	}

	go net.puller(ctx, wg)
	go net.pusher(ctx, wg)
}
