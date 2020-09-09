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

package libumi

import (
	"encoding/binary"
)

// TxStruct is ...
type TxStruct []byte

// NewTxCrtStruct ...
func NewTxCrtStruct() TxStruct {
	t := make([]byte, TxLength)
	setTxVersion(t, CreateStructure)

	return t
}

// NewTxUpdStruct ...
func NewTxUpdStruct() TxStruct {
	t := make([]byte, TxLength)
	setTxVersion(t, UpdateStructure)

	return t
}

// Version ...
func (t TxStruct) Version() uint8 {
	return t[0]
}

// Sender ...
func (t TxStruct) Sender() Address {
	return Address(t[1:35])
}

// SetSender ...
func (t TxStruct) SetSender(a Address) {
	copy(t[1:35], a)
}

// Prefix ...
func (t TxStruct) Prefix() string {
	return addressVersionToPrefix(t[35], t[36])
}

// SetPrefix ...
func (t TxStruct) SetPrefix(s string) {
	t[35], t[36] = prefixToAddressVersion(s)
}

// ProfitPercent ...
func (t TxStruct) ProfitPercent() uint16 {
	return binary.BigEndian.Uint16(t[37:39])
}

// SetProfitPercent ...
func (t TxStruct) SetProfitPercent(n uint16) {
	binary.BigEndian.PutUint16(t[37:39], n)
}

// FeePercent ...
func (t TxStruct) FeePercent() uint16 {
	return binary.BigEndian.Uint16(t[39:41])
}

// SetFeePercent ...
func (t TxStruct) SetFeePercent(p uint16) {
	binary.BigEndian.PutUint16(t[39:41], p)
}

// Name ...
func (t TxStruct) Name() string {
	return string(t[42:(42 + t[41])])
}

// SetName ...
func (t TxStruct) SetName(s string) {
	t[41] = uint8(len(s))
	copy(t[42:77], s)
}
