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

package method

import (
	"encoding/json"
	"umid/umid"
)

// ListBlocks ...
type ListBlocks struct{}

// Name ...
func (ListBlocks) Name() string {
	return "listBlocks"
}

// Process ...
func (ListBlocks) Process(bc umid.IBlockchain, params json.RawMessage) (result json.RawMessage, error json.RawMessage) {
	prm := new(struct {
		Height uint64 `json:"height"`
	})

	if err := json.Unmarshal(params, prm); err != nil {
		return nil, ErrInvalidParams
	}

	b, err := bc.BlocksByHeight(prm.Height)
	if err != nil {
		return nil, ErrInternalError
	}

	return marshalBlocks(b), nil
}

func marshalBlocks(v interface{}) json.RawMessage {
	jsn, _ := json.Marshal(v)

	return jsn
}
