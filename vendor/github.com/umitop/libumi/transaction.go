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
	"crypto/sha256"
	"encoding/binary"
)

// TxLength ...
const TxLength = 150

// Types.
const (
	Genesis uint8 = iota
	Basic
	CreateStructure
	UpdateStructure
	UpdateProfitAddress
	UpdateFeeAddress
	CreateTransitAddress
	DeleteTransitAddress
)

// Transaction ...
type Transaction []byte

// Hash ...
func (t Transaction) Hash() []byte {
	h := sha256.Sum256(t)

	return h[:]
}

// Version ...
func (t Transaction) Version() uint8 {
	return t[0]
}

// Sender ...
func (t Transaction) Sender() Address {
	return Address(t[1:35])
}

// Recipient ...
func (t Transaction) Recipient() Address {
	return Address(t[35:69])
}

// Value ...
func (t Transaction) Value() uint64 {
	return binary.BigEndian.Uint64(t[69:77])
}

// Prefix ...
func (t Transaction) Prefix() string {
	return versionToPrefix(binary.BigEndian.Uint16(t[35:37]))
}

// ProfitPercent ...
func (t Transaction) ProfitPercent() uint16 {
	return binary.BigEndian.Uint16(t[37:39])
}

// FeePercent ...
func (t Transaction) FeePercent() uint16 {
	return binary.BigEndian.Uint16(t[39:41])
}

// Name ...
func (t Transaction) Name() string {
	return string(t[42:(42 + t[41])])
}

// Nonce ...
func (t Transaction) Nonce() uint64 {
	return binary.BigEndian.Uint64(t[77:85])
}

// Signature ...
func (t Transaction) Signature() []byte {
	return t[85:149]
}

// Verify ...
func (t Transaction) Verify() error {
	return assert(t,
		lengthIs(TxLength), versionIsValid,

		ifVersionIsGenesis(
			senderPrefixIs(genesis), recipientPrefixIs(umi),
		),

		ifVersionIsBasic(
			senderAndRecipientNotEqual, senderPrefixValidAndNot(genesis), recipientPrefixValidAndNot(genesis),
		),

		ifVersionIsCreateOrUpdateStruct(
			senderPrefixIs(umi),
			structPrefixValidAndNot(genesis, umi),
			profitPercentBetween(1_00, 5_00), //nolint:gomnd
			feePercentBetween(0, 20_00),
			nameIsValid,
		),

		ifVersionIsUpdateAddress(
			senderPrefixIs(umi), recipientPrefixValidAndNot(genesis, umi),
		),

		signatureIsValid,
	)
}
