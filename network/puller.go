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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const pullIntervalSec = 5

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

	req, _ := http.NewRequestWithContext(net.ctx, "POST", peer(), strings.NewReader(jsn))

	resp, err := net.client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if err != nil {
		return
	}

	return net.parseResponse(body)
}

func (net *Network) parseResponse(body []byte) (ok bool) {
	res := &struct{ Result [][]byte }{}

	if err := json.Unmarshal(body, res); err != nil {
		return
	}

	for _, b := range res.Result {
		if err := net.blockchain.AddBlock(b); err != nil {
			return
		}
	}

	return len(res.Result) > 0
}
