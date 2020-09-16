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

// HeaderLength ...
const HeaderLength = 167

// Block ...
type Block []byte

// Hash ...
func (b Block) Hash() []byte {
	h := sha256.Sum256(b[:HeaderLength])

	return h[:]
}

// Version ...
func (b Block) Version() uint8 {
	return b[0]
}

// PreviousBlockHash ...
func (b Block) PreviousBlockHash() []byte {
	return b[1:33]
}

// MerkleRootHash ...
func (b Block) MerkleRootHash() []byte {
	return b[33:65]
}

// Timestamp ..
func (b Block) Timestamp() uint32 {
	return binary.BigEndian.Uint32(b[65:69])
}

// TxCount ...
func (b Block) TxCount() uint16 {
	return binary.BigEndian.Uint16(b[69:71])
}

// PublicKey ...
func (b Block) PublicKey() []byte {
	return b[71:103]
}

// Transaction ...
func (b Block) Transaction(idx uint16) Transaction {
	x := HeaderLength + int(idx)*TxLength
	y := x + TxLength

	return Transaction(b[x:y])
}

// Verify ...
func (b Block) Verify() error {
	return assert(b,
		lengthIsValid,
		versionIsValid,
		signatureIsValid,

		ifVersionIsGenesis(
			prevBlockHashIsNull,
			allTransactionAreGenesis,
		),

		ifVersionIsBasic(
			prevBlockHashNotNull,
			allTransactionNotGenesis,
		),

		merkleRootIsValid,
		allTransactionsAreValid,
	)
}
