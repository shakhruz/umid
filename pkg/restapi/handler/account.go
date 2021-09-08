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
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gitlab.com/umitop/umid/pkg/umi"
)

type GetAccountResponse struct {
	Data  *GetAccountData `json:"data,omitempty"`
	Error *Error          `json:"error,omitempty"`
}

type GetAccountData struct {
	Type               string `json:"type"`
	ConfirmedBalance   uint64 `json:"confirmedBalance"`
	UnconfirmedBalance int64  `json:"unconfirmedBalance"`
	TransactionCount   uint64 `json:"transactionCount"`

	Balance      *uint64 `json:"balance,omitempty"`
	InterestRate *uint16 `json:"interestRate,omitempty"`
	UpdatedAt    *string `json:"updatedAt,omitempty"`

	CompositeBalance      *uint64 `json:"compositeBalance,omitempty"`
	CompositeInterestRate *uint16 `json:"compositeInterestRate,omitempty"`
	CompositeUpdatedAt    *string `json:"compositeUpdatedAt,omitempty"`

	DeductibleBalance      *uint64 `json:"deductibleBalance,omitempty"`
	DeductibleInterestRate *uint16 `json:"deductibleInterestRate,omitempty"`
	DeductibleUpdatedAt    *string `json:"deductibleUpdatedAt,omitempty"`
}

func GetAccount(ledger1 iLedger, mempool iMempool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(GetAccountResponse)
		response.Data, response.Error = processGetAccount(r, ledger1, mempool)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func processGetAccount(r *http.Request, ledger1 iLedger, mempool iMempool) (*GetAccountData, *Error) {
	bech32 := strings.TrimPrefix(r.URL.Path, "/api/addresses/")
	bech32 = strings.TrimSuffix(bech32, "/account")

	address, err := umi.ParseAddress(bech32)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	account, ok := ledger1.Account(address)
	if !ok {
		return nil, NewError(404, "Account not found")
	}

	switch account.Type {
	case umi.Umi:
		return processUmi(address, ledger1, mempool)
	case umi.Deposit, umi.Transit, umi.Fee:
		return processDeposit(address, ledger1, mempool)
	case umi.Profit:
		return processProfit(address, ledger1, mempool)
	case umi.Dev:
		return processDev(address, ledger1, mempool)
	default:
		return nil, NewError(503, "unknown account type")
	}
}

func processUmi(address umi.Address, ledger1 iLedger, mempool iMempool) (*GetAccountData, *Error) {
	account, _ := ledger1.Account(address)

	data := &GetAccountData{
		Type:             account.Type.String(),
		ConfirmedBalance: account.BalanceAt(uint32(time.Now().Unix())),
		TransactionCount: account.TransactionCount,
	}

	data.UnconfirmedBalance = int64(data.ConfirmedBalance) + mempool.UnconfirmedBalance(address)

	return data, nil
}

func processDeposit(address umi.Address, ledger1 iLedger, mempool iMempool) (*GetAccountData, *Error) {
	account, _ := ledger1.Account(address)

	data := &GetAccountData{
		Type:             account.Type.String(),
		ConfirmedBalance: account.BalanceAt(uint32(time.Now().Unix())),
		TransactionCount: account.TransactionCount,
	}

	data.UnconfirmedBalance = int64(data.ConfirmedBalance) + mempool.UnconfirmedBalance(address)

	data.Balance = new(uint64)
	*data.Balance = account.Balance

	data.InterestRate = new(uint16)
	*data.InterestRate = account.InterestRate

	data.UpdatedAt = new(string)
	*data.UpdatedAt = time.Unix(int64(account.UpdatedAt), 0).UTC().Format(time.RFC3339)

	return data, nil
}

func processProfit(address umi.Address, ledger1 iLedger, mempool iMempool) (*GetAccountData, *Error) {
	account, _ := ledger1.Account(address)

	data := &GetAccountData{
		Type:             account.Type.String(),
		TransactionCount: account.TransactionCount,
	}

	data.CompositeBalance = new(uint64)
	*data.CompositeBalance = account.Balance

	data.CompositeInterestRate = new(uint16)
	*data.CompositeInterestRate = account.InterestRate

	data.CompositeUpdatedAt = new(string)
	*data.CompositeUpdatedAt = time.Unix(int64(account.UpdatedAt), 0).UTC().Format(time.RFC3339)

	structure, _ := ledger1.Structure(address.Prefix())

	data.DeductibleBalance = new(uint64)
	*data.DeductibleBalance = structure.Balance

	data.DeductibleInterestRate = new(uint16)
	*data.DeductibleInterestRate = structure.LevelInterestRate

	data.DeductibleUpdatedAt = new(string)
	*data.DeductibleUpdatedAt = time.Unix(int64(structure.UpdatedAt), 0).UTC().Format(time.RFC3339)

	timestamp := uint32(time.Now().Unix())
	data.ConfirmedBalance = account.BalanceAt(timestamp) - structure.BalanceAt(timestamp)
	data.UnconfirmedBalance = int64(data.ConfirmedBalance) + mempool.UnconfirmedBalance(address)

	return data, nil
}

func processDev(address umi.Address, ledger1 iLedger, mempool iMempool) (*GetAccountData, *Error) {
	account, _ := ledger1.Account(address)

	data := &GetAccountData{
		Type:             account.Type.String(),
		TransactionCount: account.TransactionCount,
	}

	data.CompositeBalance = new(uint64)
	*data.CompositeBalance = account.Balance

	data.CompositeInterestRate = new(uint16)
	*data.CompositeInterestRate = account.InterestRate

	data.CompositeUpdatedAt = new(string)
	*data.CompositeUpdatedAt = time.Unix(int64(account.UpdatedAt), 0).UTC().Format(time.RFC3339)

	structure, _ := ledger1.Structure(address.Prefix())
	profitAccount, _ := ledger1.Account(structure.ProfitAddress)

	data.DeductibleBalance = new(uint64)
	*data.DeductibleBalance = profitAccount.Balance

	data.DeductibleInterestRate = new(uint16)
	*data.DeductibleInterestRate = profitAccount.InterestRate

	data.DeductibleUpdatedAt = new(string)
	*data.DeductibleUpdatedAt = time.Unix(int64(profitAccount.UpdatedAt), 0).UTC().Format(time.RFC3339)

	timestamp := uint32(time.Now().Unix())
	data.ConfirmedBalance = account.BalanceAt(timestamp) - profitAccount.BalanceAt(timestamp)
	data.UnconfirmedBalance = int64(data.ConfirmedBalance) + mempool.UnconfirmedBalance(address)

	return data, nil
}
