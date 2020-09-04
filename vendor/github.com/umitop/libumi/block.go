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
	"errors"
)

var (
	// ErrBlkIndexOverflow ...
	ErrBlkIndexOverflow = errors.New("block: too many transactions")
	// ErrBlkInvalidSignature ...
	ErrBlkInvalidSignature = errors.New("block: invalid signature")
	// ErrBlkInvalidVersion ...
	ErrBlkInvalidVersion = errors.New("block: invalid version")
	// ErrBlkInvalidLength ...
	ErrBlkInvalidLength = errors.New("block: invalid length")
	// ErrBlkNonUniqueTrx ...
	ErrBlkNonUniqueTrx = errors.New("block: non-unique transaction")
)

// Block ...
type Block []byte

// NewBlock ...
func NewBlock() Block {
	b := make(Block, HeaderLength)
	_ = b.SetVersion(1)

	return b
}

// Hash ...
func (b Block) Hash() []byte {
	h := sha256.Sum256(b[:HeaderLength])

	return h[:]
}

// Header ...
func (b Block) Header() Header {
	return Header(b[:HeaderLength])
}

// Version ...
func (b Block) Version() uint8 {
	return b[0]
}

// SetVersion ...
func (b Block) SetVersion(ver uint8) (err error) {
	if ver > 1 {
		return ErrBlkInvalidVersion
	}

	b[0] = ver

	return err
}

// PreviousBlockHash ...
func (b Block) PreviousBlockHash() []byte {
	return b[1:33]
}

// SetPreviousBlockHash ...
func (b Block) SetPreviousBlockHash(h []byte) (err error) {
	if len(h) != 32 {
		return ErrBlkInvalidLength
	}

	copy(b[1:33], h)

	return err
}

// MerkleRootHash ...
func (b Block) MerkleRootHash() []byte {
	return b[33:65]
}

// SetMerkleRootHash ...
func (b Block) SetMerkleRootHash(h []byte) (err error) {
	if len(h) != 32 {
		return ErrBlkInvalidLength
	}

	copy(b[33:65], h)

	return err
}

// Timestamp ..
func (b Block) Timestamp() uint32 {
	return binary.BigEndian.Uint32(b[65:69])
}

// SetTimestamp ...
func (b Block) SetTimestamp(t uint32) {
	binary.BigEndian.PutUint32(b[65:69], t)
}

// TxCount ...
func (b Block) TxCount() uint16 {
	return binary.BigEndian.Uint16(b[69:71])
}

func (b Block) setTxCount(n uint16) {
	binary.BigEndian.PutUint16(b[69:71], n)
}

// PublicKey ...
func (b Block) PublicKey() ed25519.PublicKey {
	return ed25519.PublicKey(b[71:103])
}

// SetPublicKey ...
func (b Block) SetPublicKey(k ed25519.PublicKey) {
	copy(b[71:103], k)
}

// Signature ...
func (b Block) Signature() []byte {
	return b[103:167]
}

// SetSignature ...
func (b Block) SetSignature(s []byte) (err error) {
	if len(s) != ed25519.SignatureSize {
		return ErrBlkInvalidLength
	}

	copy(b[103:167], s)

	return err
}

// Sign ...
func (b Block) Sign(k ed25519.PrivateKey) {
	b.SetPublicKey(k.Public().(ed25519.PublicKey))
	_ = b.SetSignature(ed25519.Sign(k, b[:103]))
}

// Transaction ...
func (b Block) Transaction(idx uint16) Transaction {
	x := int(idx)*TransactionLength + HeaderLength
	y := x + TransactionLength

	return Transaction(b[x:y])
}

// AppendTransaction ...
func (b *Block) AppendTransaction(t Transaction) (err error) {
	blk := *b

	c := blk.TxCount()
	if c > 65534 {
		return ErrBlkIndexOverflow
	}

	blk = append(blk, t...)
	blk.setTxCount(c + 1)

	*b = blk

	return err
}

// Verify ...
func (b Block) Verify() error {
	if !ed25519.Verify(b.PublicKey(), b[0:103], b.Signature()) {
		return ErrBlkInvalidSignature
	}

	return nil
}

// CalculateMerkleRoot ...
func (b Block) CalculateMerkleRoot() (hsh []byte, err error) {
	c := b.TxCount()

	h := make([][32]byte, c)
	u := map[[32]byte]struct{}{}

	// step 1

	for i := uint16(0); i < c; i++ {
		h[i] = sha256.Sum256(b.Transaction(i))
		if _, ok := u[h[i]]; ok {
			return hsh, ErrBlkNonUniqueTrx
		}

		u[h[i]] = struct{}{}
	}

	// step 2

	min := func(a, b int) int {
		if a > b {
			return b
		}

		return a
	}

	next := func(count int) (nextCount, maxIdx int) {
		maxIdx = count - 1

		if count > 2 {
			count += count % 2
		}

		nextCount = count / 2

		return nextCount, maxIdx
	}

	t := make([]byte, 64)

	for n, m := next(int(c)); n > 0; n, m = next(n) {
		for i := 0; i < n; i++ {
			k1 := i * 2
			k2 := min(k1+1, m)
			copy(t[:32], h[k1][:])
			copy(t[32:], h[k2][:])
			h[i] = sha256.Sum256(t)
		}
	}

	hsh = make([]byte, 32)
	copy(hsh, h[0][:])

	return hsh, err
}
