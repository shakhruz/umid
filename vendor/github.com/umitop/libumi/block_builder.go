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

// BlockBuilder ...
type BlockBuilder []byte

// NewBlockBuilder ...
func NewBlockBuilder() BlockBuilder {
	b := make(BlockBuilder, HeaderLength)
	b.SetVersion(Basic)

	return b
}

// AppendTransaction ...
func (b *BlockBuilder) AppendTransaction(t []byte) {
	blk := *b

	blk = append(blk, t...)
	blk.SetTxCount(blk.TxCount() + 1)

	*b = blk
}

// SetVersion ...
func (b BlockBuilder) SetVersion(v uint8) BlockBuilder {
	b[0] = v

	return b
}

// SetPreviousBlockHash ...
func (b BlockBuilder) SetPreviousBlockHash(h []byte) BlockBuilder {
	copy(b[1:33], h)

	return b
}

// SetMerkleRootHash ...
func (b BlockBuilder) SetMerkleRootHash(h []byte) BlockBuilder {
	copy(b[33:65], h)

	return b
}

// SetTimestamp ...
func (b BlockBuilder) SetTimestamp(t uint32) BlockBuilder {
	binary.BigEndian.PutUint32(b[65:69], t)

	return b
}

// TxCount ...
func (b BlockBuilder) TxCount() uint16 {
	return binary.BigEndian.Uint16(b[69:71])
}

// SetTxCount ...
func (b BlockBuilder) SetTxCount(n uint16) BlockBuilder {
	binary.BigEndian.PutUint16(b[69:71], n)

	return b
}

// SetPublicKey ...
func (b BlockBuilder) SetPublicKey(pub []byte) BlockBuilder {
	copy(b[71:103], pub)

	return b
}

// SetSignature ...
func (b BlockBuilder) SetSignature(sig []byte) BlockBuilder {
	copy(b[103:167], sig)

	return b
}

// Sign ...
func (b BlockBuilder) Sign(sec []byte) BlockBuilder {
	mrk, _ := calculateMerkleRoot(Block(b))

	b.SetMerkleRootHash(mrk)
	b.SetPublicKey((ed25519.PrivateKey)(sec).Public().(ed25519.PublicKey))
	b.SetSignature(ed25519.Sign(sec, b[0:103]))

	return b
}

// Build ...
func (b BlockBuilder) Build() Block {
	blk := make(Block, len(b))
	copy(blk, b)

	return blk
}

// calculateMerkleRoot ...
func calculateMerkleRoot(b Block) ([]byte, error) {
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
