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

// GetStructure ...
type GetStructure struct{}

// Name ...
func (l GetStructure) Name() string {
	return "getStructure"
}

// Process ...
func (l GetStructure) Process(bc umid.IBlockchain, params json.RawMessage) (json.RawMessage, json.RawMessage) {
	prm := new(struct {
		Prefix string `json:"prefix"`
	})

	if err := json.Unmarshal(params, prm); err != nil || len(prm.Prefix) != 3 {
		return nil, ErrInvalidParams
	}

	s, err := bc.StructureByPrefix(prm.Prefix)
	if err != nil {
		return nil, ErrInternalError
	}

	return marshalStructure(s), nil
}

// ListStructures ...
type ListStructures struct{}

// Name ...
func (ListStructures) Name() string {
	return "listStructures"
}

// Process ...
func (ListStructures) Process(bc umid.IBlockchain, _ json.RawMessage) (result json.RawMessage, error json.RawMessage) {
	s, err := bc.Structures()
	if err != nil {
		return nil, ErrInternalError
	}

	arr := make([]json.RawMessage, len(s))

	for i, v := range s {
		arr[i] = marshalStructure(v)
	}

	result, _ = json.Marshal(arr)

	return
}

func marshalStructure(v interface{}) json.RawMessage {
	jsn, _ := json.Marshal(v)

	return jsn
}
