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
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"umid/jsonrpc"
)

func TestHttpMethodNotAllowed(t *testing.T) {
	ctx := context.Background()
	rpc := jsonrpc.NewRPC()

	req, _ := http.NewRequestWithContext(ctx, "GET", "/json-rpc", nil)

	res := httptest.NewRecorder()
	handler := http.HandlerFunc(jsonrpc.CORS(jsonrpc.Filter(rpc.HTTP)))
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("wrong httpCode code: got %v want %v", status, http.StatusMethodNotAllowed)
	}

	if res.Header().Get("Content-Type") != "application/json" {
		t.Errorf("wrong Content-Type header: got %v want %v", res.Header().Get("Content-Type"), "application/json")
	}

	expected := `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`
	if res.Body.String() != expected {
		t.Errorf("unexpected body: got %v want %v", res.Body.String(), expected)
	}
}

func TestHttpEntityTooLarge(t *testing.T) {
	ctx := context.Background()
	rpc := jsonrpc.NewRPC()

	req, _ := http.NewRequestWithContext(ctx, "POST", "/json-rpc", strings.NewReader(strings.Repeat(".", 11*1024*1024)))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	handler := http.HandlerFunc(jsonrpc.CORS(jsonrpc.Filter(rpc.HTTP)))
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusRequestEntityTooLarge {
		t.Errorf("wrong httpCode code: got %v want %v", status, http.StatusRequestEntityTooLarge)
	}

	if res.Header().Get("Content-Type") != "application/json" {
		t.Errorf("wrong Content-Type header: got %v want %v", res.Header().Get("Content-Type"), "application/json")
	}

	expected := `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`
	if res.Body.String() != expected {
		t.Errorf("unexpected body: got %v want %v", res.Body.String(), expected)
	}
}

func TestHttpNoContent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rpc := jsonrpc.NewRPC()
	go rpc.Worker(ctx, &sync.WaitGroup{})

	body := strings.NewReader(`{"jsonrpc":"2.0","method":"notify"}`)
	req, _ := http.NewRequestWithContext(ctx, "POST", "/json-rpc", body)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	handler := http.HandlerFunc(jsonrpc.CORS(jsonrpc.Filter(rpc.HTTP)))
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusNoContent {
		t.Errorf("wrong httpCode code: got %v want %v", status, http.StatusNoContent)
	}

	if res.Header().Get("Content-Type") != "" {
		t.Errorf("wrong Content-Type header: got %v want %v", res.Header().Get("Content-Type"), "nothing")
	}

	if res.Body.String() != "" {
		t.Errorf("unexpected body: got %v want %v", res.Body.String(), "")
	}
}

/*
func TestHttpEmptyRequest(t *testing.T) {
	ctx := context.Background()
	rpc := jsonrpc.NewRPC()
	go rpc.Worker(ctx, &sync.WaitGroup{})

	req, _ := http.NewRequestWithContext(ctx, "POST", "/json-rpc", strings.NewReader(""))
	req.Header.Set("Content-Type", "text/plain")

	res := httptest.NewRecorder()
	handler := http.HandlerFunc(jsonrpc.CORS(jsonrpc.Filter(rpc.HTTP)))
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusBadRequest {
		t.Errorf("wrong httpCode code: got %v want %v", status, http.StatusBadRequest)
	}

	if res.Header().Get("Content-Type") != "application/json" {
		t.Errorf("wrong Content-Type header: got %v want %v", res.Header().Get("Content-Type"), "application/json")
	}

	expected := `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`
	if res.Body.String() != expected {
		t.Errorf("unexpected body: got %v want %v", res.Body.String(), expected)
	}
}
*/

func TestHttpCORS(t *testing.T) {
	ctx := context.Background()
	rpc := jsonrpc.NewRPC()

	req, _ := http.NewRequestWithContext(ctx, "OPTIONS", "/json-rpc", nil)
	req.Header.Set("Origin", "*")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	res := httptest.NewRecorder()
	handler := http.HandlerFunc(jsonrpc.CORS(jsonrpc.Filter(rpc.HTTP)))
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusNoContent {
		t.Errorf("wrong httpCode code: got %v want %v", status, http.StatusNoContent)
	}

	if res.Header().Get("Content-Type") != "" {
		t.Errorf("wrong Content-Type header: got %v want %v", res.Header().Get("Content-Type"), "")
	}

	if res.Body.String() != "" {
		t.Errorf("unexpected body: got %v want %v", res.Body.String(), "")
	}
}

func TestHttpContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	rpc := jsonrpc.NewRPC()

	body := strings.NewReader(`{"jsonrpc":"2.0","method":"listStructures","id":1}`)
	req, _ := http.NewRequestWithContext(ctx, "POST", "/json-rpc", body)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	handler := http.HandlerFunc(jsonrpc.CORS(jsonrpc.Filter(rpc.HTTP)))
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusOK {
		t.Errorf("wrong httpCode code: got %v want %v", status, http.StatusOK)
	}

	if res.Header().Get("Content-Type") != "application/json" {
		t.Errorf("wrong Content-Type header: got %v want %v", res.Header().Get("Content-Type"), "application/json")
	}

	if res.Body.String() != "" {
		t.Errorf("unexpected body: got %v want %v", res.Body.String(), "")
	}
}

func TestHttpTooManyRequests(t *testing.T) {
	ctx := context.Background()
	rpc := jsonrpc.NewRPC()

	body := strings.NewReader(`{"jsonrpc":"2.0","method":"listStructures","id":1}`)
	req, _ := http.NewRequestWithContext(ctx, "POST", "/json-rpc", body)
	req.Header.Set("Content-Type", "application/json")

	for i := 0; i < 1025; i++ {
		go func() {
			res := httptest.NewRecorder()
			handler := http.HandlerFunc(jsonrpc.CORS(jsonrpc.Filter(rpc.HTTP)))
			handler.ServeHTTP(res, req)
		}()
	}

	time.Sleep(time.Second)

	res := httptest.NewRecorder()
	handler := http.HandlerFunc(jsonrpc.CORS(jsonrpc.Filter(rpc.HTTP)))
	handler.ServeHTTP(res, req)

	if status := res.Code; status != http.StatusTooManyRequests {
		t.Errorf("wrong httpCode code: got %v want %v", status, http.StatusTooManyRequests)
	}

	if res.Header().Get("Content-Type") != "application/json" {
		t.Errorf("wrong Content-Type header: got %v want %v", res.Header().Get("Content-Type"), "application/json")
	}

	expected := `{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"},"id":null}`
	if res.Body.String() != expected {
		t.Errorf("unexpected body: got %v want %v", res.Body.String(), expected)
	}
}
