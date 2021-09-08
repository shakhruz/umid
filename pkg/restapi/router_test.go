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

package restapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/restapi/handler"
	"gitlab.com/umitop/umid/pkg/umi"
)

type mockLedger struct {
	AccountFn    func(address umi.Address) (account *ledger.Account, ok bool)
	StructureFn  func(pfx umi.Prefix) (structure *ledger.Structure, ok bool)
	StructuresFn func() (structures []*ledger.Structure)
}

type mockMempool struct{}

func (mock *mockLedger) Account(address umi.Address) (account *ledger.Account, ok bool) {
	if mock.AccountFn == nil {
		return nil, false
	}

	return mock.AccountFn(address)
}

func (mock *mockLedger) Structure(pfx umi.Prefix) (structure *ledger.Structure, ok bool) {
	if mock.StructureFn == nil {
		return nil, false
	}

	return mock.StructureFn(pfx)
}

func (mock *mockLedger) Structures() (structures []*ledger.Structure) {
	if mock.StructuresFn == nil {
		return nil
	}

	return mock.StructuresFn()
}

func (mock *mockMempool) Mempool() (transactions []*umi.Transaction) {
	return nil
}

func (mock *mockMempool) Transactions(address umi.Address) (transactions []*umi.Transaction) {
	return nil
}

func (mock *mockMempool) Push(transaction umi.Transaction) error {
	return nil
}

func (mock *mockMempool) UnconfirmedBalance(address umi.Address) int64 {
	return 0
}

func TestEventsHandlerGetAccount_OK(t *testing.T) {
	t.Parallel()

	addr := "roy1y8tdlvup2ja964jwp2revprjvnmc4wku80z0eg42sftqwkzwg6vsys7kds"
	target := fmt.Sprintf("/api/addresses/%s/account", addr)

	r := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()

	account1 := &ledger.Account{
		Type:             umi.Transit,
		TransactionCount: 42,
		Balance:          123456,
		UpdatedAt:        uint32(time.Now().Unix()),
		InterestRate:     1234,
	}

	mLedger := &mockLedger{
		AccountFn: func(address umi.Address) (account *ledger.Account, ok bool) {
			return account1, true
		},
	}

	mMempool := &mockMempool{}

	handler.GetAccount(mLedger, mMempool)(w, r)

	codeGot := w.Code
	codeWant := http.StatusOK
	if codeGot != codeWant {
		t.Errorf("got status %d but wanted %d", codeGot, codeWant)
	}

	typeGot := w.Header().Get("Content-Type")
	typeWant := "application/json; charset=utf-8"
	if typeGot != typeWant {
		t.Errorf("got Content-Type '%s' but wanted '%s'", typeGot, typeWant)
	}

	resp := handler.GetAccountResponse{}

	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("JSON parsing error: %v", err)
	}

	// t.Logf("%s", w.Body.String())

	if resp.Error != nil {
		t.Errorf("response JSON must NOT contain 'error' object")
	}

	if resp.Data == nil {
		t.Fatalf("response JSON must contain 'data' object")
	}
}

func TestEventsHandlerGetAccount_NotFound(t *testing.T) {
	t.Parallel()

	addr := "roy1y8tdlvup2ja964jwp2revprjvnmc4wku80z0eg42sftqwkzwg6vsys7kds"
	target := fmt.Sprintf("/api/addresses/%s/account", addr)

	r := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()

	handler.GetAccount(&mockLedger{}, nil)(w, r)

	codeGot := w.Code
	codeWant := http.StatusOK

	if codeGot != codeWant {
		t.Errorf("got status %d but wanted %d", codeGot, codeWant)
	}

	typeGot := w.Header().Get("Content-Type")
	typeWant := "application/json; charset=utf-8"

	if typeGot != typeWant {
		t.Errorf("got Content-Type '%s' but wanted '%s'", typeGot, typeWant)
	}

	resp := handler.GetAccountResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)

	if err != nil {
		t.Fatalf("JSON parsing error: %v", err)
	}

	if resp.Data != nil {
		t.Errorf("response JSON must NOT contain 'data' object")
	}

	if resp.Error == nil {
		t.Errorf("response JSON must contain 'error' object")
	}

	if resp.Error.Code != 404 {
		t.Errorf("error.code must be 404")
	}
}

func TestEventsHandlerGetAccount_MalformedAddress(t *testing.T) {
	t.Parallel()

	addr := "qwerty"
	target := fmt.Sprintf("/api/addresses/%s/account", addr)

	r := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()

	handler.GetAccount(nil, nil)(w, r)

	codeGot := w.Code
	codeWant := http.StatusOK

	if codeGot != codeWant {
		t.Errorf("got status %d but wanted %d", codeGot, codeWant)
	}

	typeGot := w.Header().Get("Content-Type")
	typeWant := "application/json; charset=utf-8"

	if typeGot != typeWant {
		t.Errorf("got Content-Type '%s' but wanted '%s'", typeGot, typeWant)
	}

	resp := handler.GetAccountResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)

	if err != nil {
		t.Fatalf("JSON parsing error: %v", err)
	}

	if resp.Data != nil {
		t.Errorf("response JSON must NOT contain 'data' object")
	}

	if resp.Error == nil {
		t.Errorf("response JSON must contain 'error' object")
	}

	if resp.Error.Code != 400 {
		t.Errorf("error.code must be 400")
		t.Logf("%s", w.Body.String())
	}
}
