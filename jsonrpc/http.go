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
	"compress/gzip"
	"context"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	httpMaxUploadSize  = 10 * 1024 * 1024
	httpMaxRequestTime = 5
)

// HTTP ...
func (rpc *RPC) HTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req, err := readAllBody(w, r)
	if err != nil {
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

	precessResponse(ctx, res, w)
}

func precessResponse(ctx context.Context, res <-chan []byte, w http.ResponseWriter) {
	select {
	case b := <-res:
		writeResponse(b, w)
	case <-time.After(httpMaxRequestTime * time.Second):
		w.WriteHeader(http.StatusRequestTimeout)
	case <-ctx.Done():
		break
	}
}

func writeResponse(b []byte, w http.ResponseWriter) {
	if b == nil {
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusNoContent)

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func readAllBody(w http.ResponseWriter, r *http.Request) (b []byte, err error) {
	rdr := http.MaxBytesReader(w, r.Body, httpMaxUploadSize)

	if r.Header.Get("Content-Encoding") == "gzip" {
		rdr, err = gzip.NewReader(rdr)
		if err != nil {
			return nil, err
		}
	}

	b, err = ioutil.ReadAll(rdr)
	if err != nil {
		code, msg := http.StatusInternalServerError, errInternalError

		if err.Error() == "http: request body too large" {
			code, msg = http.StatusRequestEntityTooLarge, errInvalidRequest
		}

		w.WriteHeader(code)
		_, _ = w.Write(msg)

		return nil, err
	}

	return b, err
}

// CORS ...
func CORS(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// All CORS requests must have an Origin header.
		origin := r.Header.Get("Origin")
		if origin != "" {
			hdr := w.Header()
			hdr.Set("Access-Control-Allow-Credentials", "true")
			hdr.Set("Access-Control-Allow-Headers", "Content-Type")
			hdr.Set("Access-Control-Allow-Methods", "POST,OPTIONS")
			hdr.Set("Access-Control-Allow-Origin", origin)
			hdr.Set("Access-Control-Max-Age", "86400")
			hdr.Set("Vary", "Origin")

			accessControl := r.Header.Get("Access-Control-Request-Headers")
			if accessControl != "" {
				hdr.Set("Access-Control-Allow-Headers", accessControl)
			}
		}

		// CORS preflighted request
		if r.Method == "OPTIONS" {
			w.Header().Del("Content-Type")
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next(w, r)
	}
}

// Filter ...
func Filter(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write(errInvalidRequest)

			return
		}

		next(w, r)
	}
}
