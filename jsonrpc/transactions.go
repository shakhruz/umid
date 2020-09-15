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
	"encoding/binary"
	"encoding/hex"
	"encoding/json"

	"github.com/umitop/libumi"
)

type iTransaction interface {
	AddTxToMempool(raw []byte) error
	ListTxsByAddressAfterKey(adr []byte, key []byte, lim uint16) (raws [][]byte, err error)
	ListTxsByAddressBeforeKey(adr []byte, key []byte, lim uint16) (raws [][]byte, err error)
}

type listTxsParams struct {
	adr []byte
	key []byte
	lim uint16
	asc bool
}

// ListTransactions ...
func ListTransactions(bc iBlockchain, params []byte) (result []byte, errors []byte) {
	prm, errs := unmarshalTxsParams(params)
	if errs != nil {
		return nil, errs
	}

	fun := bc.ListTxsByAddressBeforeKey
	if prm.asc {
		fun = bc.ListTxsByAddressAfterKey
	}

	raws, err := fun(prm.adr, prm.key, prm.lim)
	if err != nil {
		return nil, errServiceUnavail
	}

	return marshalTxs(raws), nil
}

// SendTransaction ...
func SendTransaction(bc iBlockchain, params []byte) (result []byte, errors []byte) {
	prm := new(struct {
		RawTx []byte `json:"base64"`
	})

	if err := json.Unmarshal(params, prm); err != nil {
		return nil, errInvalidParams
	}

	if err := bc.AddTxToMempool(prm.RawTx); err != nil {
		return nil, marshalError(codeInvalidParams, err.Error())
	}

	hash := sha256.Sum256(prm.RawTx)

	result, _ = json.Marshal(struct {
		Hash string `json:"hash"`
	}{
		Hash: hex.EncodeToString(hash[:]),
	})

	return result, nil
}

func unmarshalTxsParams(params []byte) (prm listTxsParams, errs []byte) {
	var err error

	prmz := new(struct {
		Address string `json:"address"`
		Hash    string `json:"hash"`
		Limit   uint16 `json:"limit"`
		Order   string `json:"order"`
	})

	if err = json.Unmarshal(params, prmz); err != nil {
		return prm, errInvalidParams
	}

	if prm.adr, err = libumi.NewAddressFromBech32(prmz.Address); err != nil {
		return prm, marshalError(codeInvalidParams, err.Error())
	}

	if prm.key, err = hex.DecodeString(prmz.Hash); err != nil {
		return prm, errInvalidParams
	}

	prm.lim = 100
	if prmz.Limit > 0 {
		prm.lim = prmz.Limit
	}

	prm.asc = prmz.Order == "asc"

	return prm, nil
}

func marshalTxs(raws [][]byte) []byte {
	arr := make([]json.RawMessage, len(raws))
	for i, raw := range raws {
		arr[i] = marshalTx(raw)
	}

	jsn, _ := json.Marshal(arr)

	return jsn
}

func marshalTx(t tx) []byte {
	jsn, _ := json.Marshal(struct {
		Hash        string          `json:"hash"`
		Height      uint64          `json:"height"`
		ConfirmedAt uint32          `json:"confirmed_at"`
		BlockHeight uint32          `json:"block_height"`
		BlockTxIdx  uint16          `json:"block_tx_idx"`
		Version     uint8           `json:"version"`
		Sender      string          `json:"sender"`
		Recipient   string          `json:"recipient,omitempty"`
		Value       *uint64         `json:"value,omitempty"`
		FeeAddress  string          `json:"fee_address,omitempty"`
		FeeValue    uint64          `json:"fee_value,omitempty"`
		Structure   json.RawMessage `json:"structure,omitempty"`
	}{
		Hash:        t.hash(),
		Height:      t.height(),
		ConfirmedAt: t.confirmedAt(),
		BlockHeight: t.blockHeight(),
		BlockTxIdx:  t.blockTxIdx(),
		Version:     t.version(),
		Sender:      t.sender(),
		Recipient:   t.recipient(),
		Value:       t.value(),
		FeeAddress:  t.feeAddress(),
		FeeValue:    t.feeValue(),
		Structure:   t.structure(),
	})

	return jsn
}

type tx []byte

func (t tx) hash() string {
	return hex.EncodeToString(t[206:238])
}

func (t tx) height() uint64 {
	return binary.BigEndian.Uint64(t[238:246])
}

func (t tx) confirmedAt() uint32 {
	return binary.BigEndian.Uint32(t[160:164])
}

func (t tx) blockHeight() uint32 {
	return uint32(binary.BigEndian.Uint64(t[150:158]))
}

func (t tx) blockTxIdx() uint16 {
	return binary.BigEndian.Uint16(t[158:160])
}

func (t tx) version() uint8 {
	return t[0]
}

func (t tx) sender() string {
	return (libumi.Transaction)(t).Sender().Bech32()
}

func (t tx) recipient() string {
	switch t.version() {
	case libumi.CreateStructure, libumi.UpdateStructure:
		return ""
	}

	return (libumi.Transaction)(t).Recipient().Bech32()
}

func (t tx) value() (val *uint64) {
	switch t.version() {
	case libumi.Genesis, libumi.Basic, libumi.CreateStructure:
		val = new(uint64)
		*val = (libumi.Transaction)(t).Value()
	}

	return val
}

func (t tx) feeAddress() string {
	if t.feeValue() > 0 {
		return (libumi.Address)(t[172:206]).Bech32()
	}

	return ""
}

func (t tx) feeValue() uint64 {
	return binary.BigEndian.Uint64(t[164:172])
}

func (t tx) structure() []byte {
	switch t.version() {
	case libumi.CreateStructure, libumi.UpdateStructure:
		return marshalLong(t)
	case libumi.UpdateFeeAddress, libumi.UpdateProfitAddress, libumi.CreateTransitAddress, libumi.DeleteTransitAddress:
		return marshalShort(t)
	}

	return nil
}

func marshalLong(t tx) []byte {
	jsn, _ := json.Marshal(struct {
		Prefix        string `json:"prefix"`
		Name          string `json:"name"`
		ProfitPercent uint16 `json:"profit_percent"`
		FeePercent    uint16 `json:"fee_percent"`
	}{
		Prefix:        (libumi.Transaction)(t).Prefix(),
		Name:          (libumi.Transaction)(t).Name(),
		ProfitPercent: (libumi.Transaction)(t).ProfitPercent(),
		FeePercent:    (libumi.Transaction)(t).FeePercent(),
	})

	return jsn
}

func marshalShort(t tx) []byte {
	jsn, _ := json.Marshal(struct {
		Prefix string `json:"prefix"`
	}{
		Prefix: (libumi.Transaction)(t).Prefix(),
	})

	return jsn
}
