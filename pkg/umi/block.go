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
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	HdrLength = 167
)

var ErrBlock = errors.New("block")

type Block []byte

func NewBlock() Block {
	return make(Block, HdrLength)
}

func (block Block) Hash() Hash {
	return sha256.Sum256(block[:HdrLength])
}

func (block Block) Version() uint8 {
	return block[0]
}

func (block Block) SetVersion(version uint8) Block {
	block[0] = version

	return block
}

func (block Block) PreviousBlockHash() (hash Hash) {
	copy(hash[:], block[1:33])

	return hash
}

func (block Block) SetPreviousBlockHash(hash Hash) {
	copy(block[1:33], hash[:])
}

func (block Block) MerkleRootHash() (hash Hash) {
	copy(hash[:], block[33:65])

	return hash
}

func (block Block) SetMerkleRootHash(hash Hash) {
	copy(block[33:65], hash[:])
}

func (block Block) Timestamp() uint32 {
	return binary.BigEndian.Uint32(block[65:69])
}

func (block Block) SetTimestamp(timestamp uint32) {
	binary.BigEndian.PutUint32(block[65:69], timestamp)
}

func (block Block) TransactionCount() int {
	return int(binary.BigEndian.Uint16(block[69:71]))
}

func (block Block) SetTransactionCount(txCount int) Block {
	binary.BigEndian.PutUint16(block[69:71], uint16(txCount))

	return block
}

func (block Block) PublicKey() PublicKey {
	return PublicKey(block[71:103])
}

func (block Block) SetPublicKey(publicKey PublicKey) {
	copy(block[71:103], publicKey)
}

func (block Block) Transaction(idx int) Transaction {
	low := HdrLength + idx*TxConfirmedLength
	high := low + TxConfirmedLength

	return (Transaction)(block[low:high])
}

func (block Block) Verify() error {
	size := len(block)

	if size < HdrLength {
		return fmt.Errorf("%w: invalid length", ErrBlock)
	}

	if size != HdrLength+(block.TransactionCount()*TxConfirmedLength) {
		return fmt.Errorf("%w: invalid length", ErrBlock)
	}

	return nil
}

func (block Block) Legacy() BlockLegacy {
	legacyBlock := make(BlockLegacy, HdrLength, HdrLength+(TxLength*block.TransactionCount()))

	copy(legacyBlock[:HdrLength], block[:HdrLength])

	for i, j := 0, block.TransactionCount(); i < j; i++ {
		transaction := block.Transaction(i)
		legacyBlock = append(legacyBlock, transaction[0:TxLength]...)
	}

	return legacyBlock
}

func (block Block) MarshalJSON() ([]byte, error) {
	data := struct {
		Height            uint32 `json:"height"`
		Hash              string `json:"hash"`
		Version           uint8  `json:"version"`
		PreviousBlockHash string `json:"previousBlockHash"`
		Timestamp         string `json:"timestamp"`
		TransactionCount  int    `json:"transactionCount"`
	}{
		Height:            block.Transaction(0).BlockHeight(),
		Hash:              block.Hash().String(),
		Version:           block.Version(),
		PreviousBlockHash: block.PreviousBlockHash().String(),
		Timestamp:         time.Unix(int64(block.Timestamp()), 0).UTC().Format(time.RFC3339),
		TransactionCount:  block.TransactionCount(),
	}

	return json.Marshal(data) //nolint:wrapcheck // ...
}
