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

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/umi"
)

const ParamTrue = "true"

var (
	ErrOutOfRange = errors.New("out of range")
	ErrLimit      = errors.New("limit")
	ErrOffset     = errors.New("offset")
)

type iLedger interface {
	Account(address umi.Address) (account *ledger.Account, ok bool)
	Structure(pfx umi.Prefix) (structure *ledger.Structure, ok bool)
	Structures() (structures []*ledger.Structure)
}

type iMempool interface {
	Mempool() (transactions []*umi.Transaction)
	Transactions(address umi.Address) (transactions []*umi.Transaction)
	Push(transaction umi.Transaction) error
	UnconfirmedBalance(address umi.Address) int64
}

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func NewError(code int32, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 page not found", http.StatusNotFound)
	}
}

func MethodNotAllowed(allowMethod ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowMethods := []string{"OPTIONS"}
		allowMethods = append(allowMethods, allowMethod...)

		if r.Method == http.MethodOptions {
			// CORS preflight request содержит заголовки Access-Control-Request-Method и Access-Control-Request-Headers.
			if r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)

				return
			}

			w.Header().Set("Allow", strings.Join(allowMethods, ", "))
			w.WriteHeader(http.StatusOK)

			return
		}

		w.Header().Set("Allow", strings.Join(allowMethods, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	}
}

func ParseParams(r *http.Request, totalCount int) (firstIndex, lastIndex int, err error) {
	if totalCount == 0 {
		return 0, 0, nil
	}

	limit, err := parseLimit(r)
	if err != nil {
		return 0, 0, err
	}

	offset, err := parseOffset(r)
	if err != nil {
		return 0, 0, err
	}

	if offset < 0 {
		if totalCount+offset >= 0 {
			firstIndex = totalCount + offset
			lastIndex = firstIndex + limit

			if lastIndex >= totalCount {
				lastIndex = totalCount
			}

			return firstIndex, lastIndex, nil
		}

		if totalCount+offset+limit > 0 {
			firstIndex = 0
			lastIndex = totalCount + offset + limit

			return firstIndex, lastIndex, nil
		}

		return 0, 0, ErrOutOfRange
	}

	if offset >= totalCount {
		return 0, 0, ErrOutOfRange
	}

	if offset+limit >= totalCount {
		firstIndex = offset
		lastIndex = totalCount

		return firstIndex, lastIndex, nil
	}

	firstIndex = offset
	lastIndex = firstIndex + limit

	return firstIndex, lastIndex, nil
}

func parseLimit(r *http.Request) (limit int, err error) {
	limit = 10

	str := r.URL.Query().Get("limit")
	if str != "" {
		value, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return 10, fmt.Errorf("%w: %s", ErrLimit, err.Error())
		}

		limit = int(value)
	}

	if limit > 65535 {
		return 10, fmt.Errorf("%w: must be between 0 and 65535, got %d", ErrLimit, limit)
	}

	return limit, nil
}

func parseOffset(r *http.Request) (offset int, err error) {
	str := r.URL.Query().Get("offset")
	if str != "" {
		value, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %s", ErrOffset, err.Error())
		}

		offset = int(value)
	}

	return offset, nil
}

func setHeaders(w http.ResponseWriter, r *http.Request) {
	cors(w, r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
}

func cors(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
	}
}
