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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"umid/umid"
)

func sendTransaction(bc umid.IBlockchain, raw json.RawMessage, res *response) {
	prm := new(struct {
		Base64 []byte `json:"base64"`
	})

	if err := json.Unmarshal(raw, prm); err != nil || len(prm.Base64) != 150 {
		res.Error = errInvalidParams

		return
	}

	if err := bc.AddTransaction(prm.Base64); err != nil {
		res.Error = &respError{
			Code:    -1,
			Message: err.Error(),
		}

		return
	}

	hash := sha256.Sum256(prm.Base64)

	res.Result = struct {
		Hash string `json:"hash"`
	}{
		Hash: hex.EncodeToString(hash[:]),
	}
}
