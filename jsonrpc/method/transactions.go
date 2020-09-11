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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"umid/umid"
)

// ListTxs ...
type ListTxs struct{}

// Name ...
func (ListTxs) Name() string {
	return "listTransactions"
}

// Process ...
func (ListTxs) Process(bc umid.IBlockchain, params json.RawMessage) (result json.RawMessage, error json.RawMessage) {
	prm := new(struct {
		Address string `json:"address"`
	})

	if err := json.Unmarshal(params, prm); err != nil || prm.Address == "" {
		return nil, ErrInvalidParams
	}

	txs, err := bc.TransactionsByAddress(prm.Address)
	if err != nil {
		return nil, ErrInternalError
	}

	return marshalTxs(txs), nil
}

func marshalTxs(v interface{}) json.RawMessage {
	jsn, _ := json.Marshal(v)

	return jsn
}

// SendTx ...
type SendTx struct{}

// Name ...
func (SendTx) Name() string {
	return "sendTransaction"
}

// Process ...
func (SendTx) Process(bc umid.IBlockchain, params json.RawMessage) (result json.RawMessage, error json.RawMessage) {
	prm := new(struct {
		Tx []byte `json:"base64"`
	})

	if err := json.Unmarshal(params, prm); err != nil {
		return nil, ErrInvalidParams
	}

	if err := bc.AddTransaction(prm.Tx); err != nil {
		return nil, marshalError(codeInvalidParams, err.Error())
	}

	hash := sha256.Sum256(prm.Tx)

	b, _ := json.Marshal(struct {
		Hash string `json:"hash"`
	}{
		Hash: hex.EncodeToString(hash[:]),
	})

	return b, nil
}
