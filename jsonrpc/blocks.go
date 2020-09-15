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
	"encoding/json"
	"log"
)

type iBlock interface {
	ListBlocksAfterKey(key uint64, lim uint16) (raws [][]byte, err error)
}

// ListBlocks ...
func ListBlocks(bc iBlockchain, params []byte) (result []byte, errors []byte) {
	key, lim, errs := unmarshalBlocksParams(params)
	if errs != nil {
		return nil, errs
	}

	raws, err := bc.ListBlocksAfterKey(key, lim)
	if err != nil {
		log.Print(err.Error())

		return nil, errServiceUnavail
	}

	return marshalBlocks(raws), nil
}

func unmarshalBlocksParams(params []byte) (key uint64, lim uint16, errs []byte) {
	var err error

	prm := new(struct {
		Hash   string `json:"hash"`
		Limit  uint16 `json:"limit"`
		Height uint64 `json:"height"`
	})

	if err = json.Unmarshal(params, prm); err != nil {
		return 0, 0, errInvalidParams
	}

	key = prm.Height

	// if key, err = hex.DecodeString(prm.Hash); err != nil {
	// 	 return 0, 0, errInvalidParams
	// }

	lim = 5000
	if prm.Limit > 0 {
		lim = prm.Limit
	}

	return key, lim, errs
}

func marshalBlocks(raws [][]byte) []byte {
	jsn, _ := json.Marshal(raws)

	return jsn
}
