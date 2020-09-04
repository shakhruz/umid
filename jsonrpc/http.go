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

package jsonrpc

import (
	"io/ioutil"
	"net/http"
	"time"
)

const httpMaxUploadSize = 10 * 1024 * 1024

// HTTP ...
func (rpc *RPC) HTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	setCORSHeaders(w, r)

	if !validateRequest(w, r) {
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(errInvalidRequest)

		return
	}

	req, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, httpMaxUploadSize))
	if err != nil {
		code, msg := http.StatusInternalServerError, errInternalError

		if err.Error() == "http: request body too large" {
			code, msg = http.StatusRequestEntityTooLarge, errInvalidRequest
		}

		w.WriteHeader(code)
		_, _ = w.Write(msg)

		return
	}

	ctx := r.Context()
	res := make(chan []byte, 1)

	select {
	case rpc.queue <- rawRequest{ctx, req, res}:
		break
	default:
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write(errInternalError)

		return
	}

	select {
	case b := <-res:
		if b == nil {
			w.Header().Del("Content-Type")
			w.WriteHeader(http.StatusNoContent)

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	case <-time.After(time.Second):
		w.WriteHeader(http.StatusRequestTimeout)
		_, _ = w.Write([]byte(`time is up`))
	case <-ctx.Done():
		return
	}
}

func validateRequest(w http.ResponseWriter, r *http.Request) bool {
	// CORS preflighted request
	if r.Method == "OPTIONS" {
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent)

		return false
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write(errInvalidRequest)

		return false
	}

	// if r.Header.Get("Content-Type") != "application/json" {
	//	w.WriteHeader(http.StatusUnsupportedMediaType)
	//	_, _ = w.Write(errInvalidRequest)
	//
	//	return false
	// }

	return true
}

func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	// All CORS requests must have an Origin header.
	if r.Header.Get("Origin") != "" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Max-Age", "1") // 86400
		w.Header().Set("Vary", "Origin")

		if r.Header.Get("Access-Control-Request-Headers") != "" {
			w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
		}
	}
}
