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
	"strconv"
	"strings"

	"gitlab.com/umitop/umid/pkg/storage"
	"gitlab.com/umitop/umid/pkg/umi"
)

type ListTransactionsResponse struct {
	Data  *ListTransactionsData `json:"data,omitempty"`
	Error *Error                `json:"error,omitempty"`
}

type ListTransactionsData struct {
	TotalCount int               `json:"totalCount"`
	Items      []umi.Transaction `json:"items"`
}

type ListTransactionsRawResponse struct {
	Data  *ListTransactionsRawData `json:"data,omitempty"`
	Error *Error                   `json:"error,omitempty"`
}

type ListTransactionsRawData struct {
	TotalCount int      `json:"totalCount"`
	Items      [][]byte `json:"items"`
}

func ListTransactionsByAddress(blockchain storage.IBlockchain, index *storage.Index) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(ListTransactionsResponse)
		response.Data, response.Error = processListTransactionsByAddress(r, blockchain, index)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func ListTransactionsByBlock(blockchain storage.IBlockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		switch r.URL.Query().Get("raw") {
		case ParamTrue:
			response := new(ListTransactionsRawResponse)
			response.Data, response.Error = processListTxsByBlockRaw(r, blockchain)

			_ = json.NewEncoder(w).Encode(response)
		default:
			response := new(ListTransactionsResponse)
			response.Data, response.Error = processListTransactionsByBlock(r, blockchain)

			_ = json.NewEncoder(w).Encode(response)
		}
	}
}

func processListTransactionsByAddress(
	r *http.Request, blockchain storage.IBlockchain, index *storage.Index) (*ListTransactionsData, *Error) {
	bech32 := strings.TrimPrefix(r.URL.Path, "/api/addresses/")
	bech32 = strings.TrimSuffix(bech32, "/transactions")

	address, err := umi.ParseAddress(bech32)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	txs, ok := index.TransactionsByAddress(address)
	if !ok {
		return nil, NewError(404, "Not Found")
	}

	totalCount := len(*txs)

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	txz := (*txs)[firstIndex:lastIndex]
	transactions := make([]umi.Transaction, 0, len(txz))

	for _, key := range txz {
		blockHeight := uint32(key >> 16)
		txIndex := uint16(key & 0xFFFF)

		transaction, ok := blockchain.Transaction(blockHeight, txIndex)
		if !ok {
			return nil, NewError(503, "Internal error")
		}

		transactions = append(transactions, transaction)
	}

	data := &ListTransactionsData{
		TotalCount: totalCount,
		Items:      transactions,
	}

	return data, nil
}

func processListTransactionsByBlock(r *http.Request, blockchain storage.IBlockchain) (*ListTransactionsData, *Error) {
	height := strings.TrimPrefix(r.URL.Path, "/api/blocks/")
	height = strings.TrimSuffix(height, "/transactions")

	blockHeight, err := strconv.ParseUint(height, 10, 32)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	block, err := blockchain.Block(uint32(blockHeight))
	if err != nil {
		return nil, NewError(404, err.Error())
	}

	totalCount := block.TransactionCount()

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	transactions := make([]umi.Transaction, 0, lastIndex-firstIndex+1)

	for i := firstIndex; i < lastIndex; i++ {
		transactions = append(transactions, block.Transaction(i))
	}

	data := &ListTransactionsData{
		TotalCount: totalCount,
		Items:      transactions,
	}

	return data, nil
}

func processListTxsByBlockRaw(r *http.Request, blockchain storage.IBlockchain) (*ListTransactionsRawData, *Error) {
	height := strings.TrimPrefix(r.URL.Path, "/api/blocks/")
	height = strings.TrimSuffix(height, "/transactions")

	blockHeight, err := strconv.ParseUint(height, 10, 32)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	block, err := blockchain.Block(uint32(blockHeight))
	if err != nil {
		return nil, NewError(404, err.Error())
	}

	totalCount := block.TransactionCount()

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	transactions := make([][]byte, 0, lastIndex-firstIndex+1)

	for i := firstIndex; i < lastIndex; i++ {
		transactions = append(transactions, block.Transaction(i))
	}

	data := &ListTransactionsRawData{
		TotalCount: totalCount,
		Items:      transactions,
	}

	return data, nil
}
