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

package jsonrpc_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"umid/jsonrpc"
	"umid/umid"
)

func TestListStructures(t *testing.T) {
	var counter int

	bc := &bcMock{}
	bc.FnStructures = func() ([]umid.Structure, error) {
		if counter == 0 {
			counter = 1
			return nil, errors.New("not found")
		}

		arr := make([]umid.Structure, 1)
		arr[0] = umid.Structure{
			Prefix: "umi",
			Name:   "UMI",
		}

		return arr, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rpc := jsonrpc.NewRPC(ctx, &sync.WaitGroup{}, bc)
	go rpc.Worker()

	tests := []struct {
		request  string
		response string
	}{
		{
			`{"jsonrpc":"2.0","method":"listStructures","id":1}`,
			`{"jsonrpc":"2.0","error":{"code":-1,"message":"not found"},"id":1}`,
		},
		{
			`{"jsonrpc":"2.0","method":"listStructures","id":2}`,
			`{"jsonrpc":"2.0","result":[{"prefix":"umi","name":"UMI","fee_percent":0,"profit_percent":0,"deposit_percent":0,"fee_address":"","profit_address":"","master_address":"","transit_addresses":null,"balance":0}],"id":2}`,
		},
	}

	for _, test := range tests {
		req, _ := http.NewRequestWithContext(ctx, "POST", "/json-rpc", strings.NewReader(test.request))
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()
		handler := http.HandlerFunc(rpc.HTTP)
		handler.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("wrong http code: got %v want %v", res.Code, http.StatusOK)
		}

		if res.Header().Get("Content-Type") != "application/json" {
			t.Errorf("wrong Content-Type header: got %v want %v", res.Header().Get("Content-Type"), "application/json")
		}

		if res.Body.String() != test.response {
			t.Errorf("unexpected body: got %v want %v", res.Body.String(), test.response)
		}
	}
}
