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
	"errors"
	"fmt"
	"gitlab.com/umitop/umid/pkg/nft"
	"net/http"
	"strings"
	"time"

	"gitlab.com/umitop/umid/pkg/umi"
)

var (
	errTimestampFuture     = errors.New("некорректная метка времени: транзакция из будущего")
	errTimestampPast       = errors.New("некорректная метка времени: просроченная траназкция")
	errProhibitedSender    = errors.New("некорректный отправитель")
	errProhibitedRecipient = errors.New("некорректный получатель")
)

type PushMempoolResponse struct {
	Data  *umi.Transaction `json:"data,omitempty"`
	Error *Error           `json:"error,omitempty"`
}

type ListMempoolResponse struct {
	Data  *ListMempoolData `json:"data,omitempty"`
	Error *Error           `json:"error,omitempty"`
}

type ListMempoolData struct {
	TotalCount int                `json:"totalCount"`
	Items      []*umi.Transaction `json:"items"`
}

type ListMempoolRawResponse struct {
	Data  *ListMempoolRawData `json:"data,omitempty"`
	Error *Error              `json:"error,omitempty"`
}

type ListMempoolRawData struct {
	TotalCount int      `json:"totalCount"`
	Items      [][]byte `json:"items"`
}

func ListMempool(mempool iMempool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		switch r.URL.Query().Get("raw") {
		case ParamTrue:
			response := new(ListMempoolRawResponse)
			response.Data, response.Error = processListMempoolRaw(r, mempool)

			_ = json.NewEncoder(w).Encode(response)

		default:
			response := new(ListMempoolResponse)
			response.Data, response.Error = processListMempool(r, mempool)

			_ = json.NewEncoder(w).Encode(response)
		}
	}
}

func ListMempoolByAddress(mempool iMempool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(ListMempoolResponse)
		response.Data, response.Error = processListMempoolByAddress(r, mempool)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func PushMempool(mempool iMempool, nftMempool iNftMempool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(PushMempoolResponse)
		response.Data, response.Error = processPushMempool(r, mempool, nftMempool)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func processListMempool(r *http.Request, mempool iMempool) (*ListMempoolData, *Error) {
	transactions := mempool.Mempool()
	totalCount := len(transactions)

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	data := &ListMempoolData{
		TotalCount: totalCount,
		Items:      transactions[firstIndex:lastIndex],
	}

	return data, nil
}

func processListMempoolRaw(r *http.Request, mempool iMempool) (*ListMempoolRawData, *Error) {
	transactions := mempool.Mempool()
	totalCount := len(transactions)

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	items := make([][]byte, 0, lastIndex-firstIndex+1)

	for _, transaction := range transactions[firstIndex:lastIndex] {
		items = append(items, *transaction)
	}

	data := &ListMempoolRawData{
		TotalCount: totalCount,
		Items:      items,
	}

	return data, nil
}

func processListMempoolByAddress(r *http.Request, mempool iMempool) (*ListMempoolData, *Error) {
	bech32 := strings.TrimPrefix(r.URL.Path, "/api/addresses/")
	bech32 = strings.TrimSuffix(bech32, "/mempool")

	address, err := umi.ParseAddress(bech32)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	transactions := mempool.Transactions(address)
	totalCount := len(transactions)

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	data := &ListMempoolData{
		TotalCount: totalCount,
		Items:      transactions[firstIndex:lastIndex],
	}

	return data, nil
}

func processPushMempool(r *http.Request, mempool iMempool, nftMempool iNftMempool) (*umi.Transaction, *Error) {
	contentType := r.Header.Get("Content-Type")

	if !strings.HasPrefix(contentType, "application/json") {
		return nil, NewError(400, "'Content-Type' must be 'application/json'")
	}

	request := struct {
		Data []byte `json:"data"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, NewError(400, err.Error())
	}

	if len(request.Data) > 1 && request.Data[0] == umi.TxV17MintNft {
		if err := TxValidateNft(request.Data); err != nil {
			return nil, NewError(400, err.Error())
		}

		if err := nftMempool.Push(request.Data); err != nil {
			return nil, NewError(400, err.Error())
		}

		return (*umi.Transaction)(&request.Data), nil
	}

	if len(request.Data) != umi.TxLength {
		return nil, NewError(400, "Malformed transaction")
	}

	transaction := (umi.Transaction)(request.Data)
	txVer := transaction.Version()

	if txVer < umi.TxV8Send || txVer > umi.TxV16Issue {
		return nil, NewError(400, "Unsupported tx version")
	}

	if err := transaction.Verify(); err != nil {
		return nil, NewError(400, err.Error())
	}

	if err := TxValidate(transaction); err != nil {
		return nil, NewError(400, err.Error())
	}

	if err := mempool.Push(transaction); err != nil {
		return nil, NewError(400, err.Error())
	}

	return &transaction, nil
}

func TxValidate(transaction umi.Transaction) error {
	currentTime := uint32(time.Now().Unix())
	txTime := transaction.Timestamp()

	if txTime > currentTime {
		return errTimestampFuture
	}

	if currentTime-txTime > 3600 {
		return errTimestampPast
	}

	if transaction.Version() != umi.TxV8Send {
		return nil
	}

	senderPrefix := transaction.Sender().Prefix()
	recipientPrefix := transaction.Recipient().Prefix()

	switch senderPrefix {
	case umi.PfxVerGls, umi.PfxVerGlz:
		switch recipientPrefix {
		case umi.PfxVerGls, umi.PfxVerGlz:
			return nil
		}

		return fmt.Errorf("%w: с адреса '%s' можно отправить только на адрес 'gls' и 'glz'",
			errProhibitedRecipient, senderPrefix.String())
	}

	if recipientPrefix == umi.PfxVerUmi {
		return nil
	}

	return fmt.Errorf("%w: с адреса '%s' можно отправить только на адрес 'umi'",
		errProhibitedRecipient, senderPrefix.String())
}

func TxValidateNft(transaction []byte) error {
	tx := (nft.Transaction)(transaction)

	currentTime := uint32(time.Now().Unix())
	txTime := tx.Timestamp()

	if txTime > currentTime {
		return errTimestampFuture
	}

	if currentTime-txTime > 3600 {
		return errTimestampPast
	}

	senderPrefix := tx.Sender().Prefix()

	if senderPrefix != umi.PfxVerNft {
		return fmt.Errorf("%w: адрес с префиксом %s не может создать NFT-токен",
			errProhibitedSender, senderPrefix.String())
	}

	return nil
}
