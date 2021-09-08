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

	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/umi"
)

type GetStructureResponse struct {
	Data  *ledger.Structure `json:"data,omitempty"`
	Error *Error            `json:"error,omitempty"`
}

type ListStructuresResponse struct {
	Data  *ListStructuresData `json:"data,omitempty"`
	Error *Error              `json:"error,omitempty"`
}

type ListStructuresData struct {
	TotalCount int                 `json:"totalCount"`
	Items      []*ledger.Structure `json:"items"`
}

func GetStructure(ledger1 iLedger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(GetStructureResponse)
		response.Data, response.Error = processGetStructure(r, ledger1)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func ListStructures(ledger1 iLedger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(ListStructuresResponse)
		response.Data, response.Error = processListStructures(r, ledger1)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func processGetStructure(r *http.Request, ledger1 iLedger) (*ledger.Structure, *Error) {
	hrp := strings.TrimPrefix(r.URL.Path, "/api/structures/")
	if len(hrp) != 3 {
		return nil, NewError(404, "Not found")
	}

	prefix := umi.ParsePrefix(hrp)

	structure, ok := ledger1.Structure(prefix)
	if !ok {
		return nil, NewError(404, "Not found")
	}

	return structure, nil
}

func processListStructures(r *http.Request, ledger1 iLedger) (*ListStructuresData, *Error) {
	structures := ledger1.Structures()
	totalCount := len(structures)

	firstIndex, lastIndex, err := ParseParams(r, totalCount)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	data := &ListStructuresData{
		TotalCount: totalCount,
		Items:      structures[firstIndex:lastIndex],
	}

	return data, nil
}
