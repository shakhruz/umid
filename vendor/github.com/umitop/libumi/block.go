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
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
)

// HeaderLength ...
const HeaderLength = 167

// Block ...
type Block []byte

// NewBlock ...
func NewBlock() Block {
	b := make(Block, HeaderLength)
	setBlockVersion(b, Basic)

	return b
}

// Hash ...
func (b Block) Hash() []byte {
	h := sha256.Sum256(b[:HeaderLength])

	return h[:]
}

// Version ...
func (b Block) Version() uint8 {
	return b[0]
}

// SetVersion ...
func (b Block) SetVersion(v uint8) Block {
	b[0] = v

	return b
}

// PreviousBlockHash ...
func (b Block) PreviousBlockHash() []byte {
	return b[1:33]
}

// SetPreviousBlockHash ...
func (b Block) SetPreviousBlockHash(h []byte) Block {
	copy(b[1:33], h)

	return b
}

// MerkleRootHash ...
func (b Block) MerkleRootHash() []byte {
	return b[33:65]
}

// SetMerkleRootHash ...
func (b Block) SetMerkleRootHash(h []byte) Block {
	copy(b[33:65], h)

	return b
}

// Timestamp ..
func (b Block) Timestamp() uint32 {
	return binary.BigEndian.Uint32(b[65:69])
}

// SetTimestamp ...
func (b Block) SetTimestamp(t uint32) Block {
	binary.BigEndian.PutUint32(b[65:69], t)

	return b
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
func (b Block) Transaction(idx uint16) []byte {
	x := HeaderLength + int(idx)*TxLength
	y := x + TxLength

	return b[x:y]
}

// AppendTransaction ...
func (b *Block) AppendTransaction(t []byte) {
	blk := *b

	blk = append(blk, t...)
	setBlockTxCount(blk, blk.TxCount()+1)

	*b = blk
}

// SignBlock ...
func SignBlock(blk []byte, sec []byte) {
	setBlockPublicKey(blk, (ed25519.PrivateKey)(sec).Public().(ed25519.PublicKey))
	setBlockSignature(blk, ed25519.Sign(sec, blk[0:103]))
}

// VerifyBlock ...
func VerifyBlock(b []byte) error {
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

// CalculateMerkleRoot ...
func CalculateMerkleRoot(b Block) ([]byte, error) {
	curCount := b.TxCount()
	h := make([][32]byte, curCount)

	if !checkUniqTx(b, h) {
		return nil, ErrNonUniqueTx
	}

	for nextCount, prevCount := nextLevel(int(curCount)); nextCount > 0; nextCount, prevCount = nextLevel(nextCount) {
		calculateLevel(h, nextCount, prevCount)
	}

	return h[0][:], nil
}

func checkUniqTx(b Block, h [][32]byte) bool {
	u := make(map[[32]byte]struct{}, b.TxCount())

	for i, l := uint16(0), b.TxCount(); i < l; i++ {
		h[i] = sha256.Sum256(b.Transaction(i))
		if _, ok := u[h[i]]; ok {
			return false
		}

		u[h[i]] = struct{}{}
	}

	return true
}

func calculateLevel(h [][32]byte, nextCount int, prevCount int) {
	t := make([]byte, 64)

	for i := 0; i < nextCount; i++ {
		k1 := i * 2 //nolint:gomnd
		k2 := min(k1+1, prevCount)
		copy(t[:32], h[k1][:])
		copy(t[32:], h[k2][:])
		h[i] = sha256.Sum256(t)
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}

	return a
}

func nextLevel(curCount int) (nextCount, maxIdx int) {
	maxIdx = curCount - 1

	if curCount > 2 { //nolint:gomnd
		curCount += curCount & 1
	}

	nextCount = curCount / 2 //nolint:gomnd

	return nextCount, maxIdx
}

func setBlockVersion(b []byte, n uint8) {
	b[0] = n
}

func setBlockPublicKey(b []byte, pub []byte) {
	copy(b[71:103], pub)
}

func setBlockSignature(b []byte, sig []byte) {
	copy(b[103:167], sig)
}

func setBlockTxCount(b []byte, n uint16) {
	binary.BigEndian.PutUint16(b[69:71], n)
}
