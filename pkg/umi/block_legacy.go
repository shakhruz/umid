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

package umi

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type BlockLegacy []byte

func NewBlockLegacy() BlockLegacy {
	return make(BlockLegacy, HdrLength)
}

func (block BlockLegacy) Hash() Hash {
	return sha256.Sum256(block[:HdrLength])
}

func (block BlockLegacy) Version() uint8 {
	return block[0]
}

func (block BlockLegacy) SetVersion(version uint8) {
	block[0] = version
}

func (block BlockLegacy) PreviousBlockHash() (hash Hash) {
	copy(hash[:], block[1:33])

	return hash
}

func (block BlockLegacy) MerkleRootHash() (hash Hash) {
	copy(hash[:], block[33:65])

	return hash
}

func (block BlockLegacy) Timestamp() uint32 {
	return binary.BigEndian.Uint32(block[65:69])
}

func (block BlockLegacy) TransactionCount() int {
	return int(binary.BigEndian.Uint16(block[69:71]))
}

func (block BlockLegacy) PublicKey() PublicKey {
	return PublicKey(block[71:103])
}

func (block BlockLegacy) Transaction(idx int) Transaction {
	transaction := make(Transaction, TxConfirmedLength)

	low := HdrLength + idx*TxLength
	high := low + TxLength

	copy(transaction[:TxLength], block[low:high])

	return transaction
}

func (block BlockLegacy) Verify() error {
	size := len(block)

	if size < HdrLength {
		return fmt.Errorf("%w: invalid length", ErrBlock)
	}

	if size != HdrLength+(block.TransactionCount()*TxLength) {
		return fmt.Errorf("%w: invalid length", ErrBlock)
	}

	return nil
}
