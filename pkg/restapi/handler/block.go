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

type GetBlockResponse struct {
	Data  *umi.Block `json:"data,omitempty"`
	Error *Error     `json:"error,omitempty"`
}

type ListBlocksResponse struct {
	Data  *ListBlocksData `json:"data,omitempty"`
	Error *Error          `json:"error,omitempty"`
}

type ListBlocksData struct {
	TotalCount int         `json:"totalCount"`
	Items      []umi.Block `json:"items"`
}

type ListBlocksRawResponse struct {
	Data  *ListBlocksRawData `json:"data,omitempty"`
	Error *Error             `json:"error,omitempty"`
}

type ListBlocksRawData struct {
	TotalCount int      `json:"totalCount"`
	Items      [][]byte `json:"items"`
}

func GetBlock(blockchain storage.IBlockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(GetBlockResponse)
		response.Data, response.Error = processGetBlock(r, blockchain)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func ListBlocks(blockchain storage.IBlockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		switch r.URL.Query().Get("raw") {
		case ParamTrue:
			response := new(ListBlocksRawResponse)
			response.Data, response.Error = processListBlocksRaw(r, blockchain)

			_ = json.NewEncoder(w).Encode(response)

		default:
			response := new(ListBlocksResponse)
			response.Data, response.Error = processListBlocks(r, blockchain)

			_ = json.NewEncoder(w).Encode(response)
		}
	}
}

func processGetBlock(r *http.Request, blockchain storage.IBlockchain) (*umi.Block, *Error) {
	height := strings.TrimPrefix(r.URL.Path, "/api/blocks/")

	blockHeight, err := strconv.ParseUint(height, 10, 32)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	block, err := blockchain.Block(uint32(blockHeight))
	if err != nil {
		return nil, NewError(404, err.Error())
	}

	return &block, nil
}

func processListBlocks(r *http.Request, blockchain storage.IBlockchain) (*ListBlocksData, *Error) {
	totalCount := blockchain.Height()

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	blocks := make([]umi.Block, 0, lastIndex-firstIndex+1)

	for i := firstIndex; i < lastIndex; i++ {
		height := uint32(i) + 1

		block, err := blockchain.Block(height)
		if err != nil {
			return nil, NewError(-1, err.Error())
		}

		blocks = append(blocks, block)
	}

	data := &ListBlocksData{
		TotalCount: totalCount,
		Items:      blocks,
	}

	return data, nil
}

func processListBlocksRaw(r *http.Request, blockchain storage.IBlockchain) (*ListBlocksRawData, *Error) {
	totalCount := blockchain.Height()

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	blocks := make([][]byte, 0, lastIndex-firstIndex+1)

	for i := firstIndex; i < lastIndex; i++ {
		height := uint32(i) + 1

		block, err := blockchain.Block(height)
		if err != nil {
			return nil, NewError(-1, err.Error())
		}

		blocks = append(blocks, block.Legacy())
	}

	data := &ListBlocksRawData{
		TotalCount: totalCount,
		Items:      blocks,
	}

	return data, nil
}
