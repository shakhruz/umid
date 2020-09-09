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
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"runtime"
	"sync"
)

// Errors.
var (
	ErrBlkInvalidLength    = errors.New("invalid length")
	ErrBlkInvalidVersion   = errors.New("invalid version")
	ErrBlkInvalidSignature = errors.New("invalid signature")
	ErrBlkInvalidPrevHash  = errors.New("invalid previous block hash")
	ErrBlkInvalidMerkle    = errors.New("invalid merkle root")
	ErrBlkInvalidTx        = errors.New("invalid transaction")
	ErrBlkNonUniqueTx      = errors.New("non-unique transaction")
)

// HeaderLength ...
const HeaderLength = 167

// Block ...
type Block []byte

// NewBlock ...
func NewBlock() Block {
	b := make(Block, HeaderLength)
	setBlkVersion(b, Basic)

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

// PreviousBlockHash ...
func (b Block) PreviousBlockHash() []byte {
	return b[1:33]
}

// SetPreviousBlockHash ...
func (b Block) SetPreviousBlockHash(h []byte) {
	copy(b[1:33], h)
}

// MerkleRootHash ...
func (b Block) MerkleRootHash() []byte {
	return b[33:65]
}

// SetMerkleRootHash ...
func (b Block) SetMerkleRootHash(h []byte) {
	copy(b[33:65], h)
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
	setBlkTxCount(blk, blk.TxCount()+1)

	*b = blk
}

// SignBlock ...
func SignBlock(blk []byte, sec []byte) {
	setBlkPublicKey(blk, (ed25519.PrivateKey)(sec).Public().(ed25519.PublicKey))
	setBlkSignature(blk, ed25519.Sign(sec, blk[0:103]))
}

// VerifyBlock ...
func VerifyBlock(b Block) error {
	return verifyBlkLength(b)
}

// CalculateMerkleRoot ...
func CalculateMerkleRoot(b Block) ([]byte, error) {
	c := b.TxCount()
	h := make([][32]byte, c)

	if !checkUniq(b, h) {
		return nil, ErrBlkNonUniqueTx
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

	return h[0][:], nil
}

func checkUniq(b Block, h [][32]byte) bool {
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

func min(a, b int) int {
	if a > b {
		return b
	}

	return a
}

func next(count int) (nextCount, maxIdx int) {
	maxIdx = count - 1

	if count > 2 {
		count += count % 2
	}

	nextCount = count / 2

	return nextCount, maxIdx
}

func setBlkVersion(b []byte, n uint8) {
	b[0] = n
}

func setBlkPublicKey(b []byte, pub []byte) {
	copy(b[71:103], pub)
}

func setBlkSignature(b []byte, sig []byte) {
	copy(b[103:167], sig)
}

func setBlkTxCount(b []byte, n uint16) {
	binary.BigEndian.PutUint16(b[69:71], n)
}

func verifyBlkLength(b Block) error {
	if len(b) < (HeaderLength + TxLength) {
		return ErrBlkInvalidLength
	}

	return verifyBlkVersion(b)
}

func verifyBlkVersion(b Block) error {
	switch b.Version() {
	case Genesis:
		return verifyBlkGenesis(b)
	case Basic:
		return verifyBlkBasic(b)
	}

	return ErrBlkInvalidVersion
}

func verifyBlkGenesis(b Block) error {
	if !bytes.Equal(b.PreviousBlockHash(), make([]byte, 32)) {
		return ErrBlkInvalidPrevHash
	}

	return verifyBlkSignature(b)
}

func verifyBlkBasic(b Block) error {
	if bytes.Equal(b.PreviousBlockHash(), make([]byte, 32)) {
		return ErrBlkInvalidPrevHash
	}

	return verifyBlkSignature(b)
}

func verifyBlkSignature(b []byte) error {
	if !ed25519.Verify(b[71:103], b[0:103], b[103:167]) {
		return ErrBlkInvalidSignature
	}

	return verifyBlkMerkleRoot(b)
}

func verifyBlkMerkleRoot(b Block) error {
	mrk, err := CalculateMerkleRoot(b)
	if err != nil {
		return err
	}

	if !bytes.Equal(b.MerkleRootHash(), mrk) {
		return ErrBlkInvalidMerkle
	}

	return verifyBlkTxs(b)
}

func verifyBlkTxs(b Block) (err error) {
	c := make(chan []byte, b.TxCount())

	for i, l := uint16(0), b.TxCount(); i < l; i++ {
		c <- b.Transaction(i)
	}

	close(c)

	runParallel(func() {
		for tx := range c {
			if !verifyBlkTxVersion(b[0], tx[0]) || VerifyTx(tx) != nil {
				err = ErrBlkInvalidTx

				return
			}
		}
	})

	return err
}

func runParallel(f func()) {
	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		go func() {
			f()
			wg.Done()
		}()
	}

	wg.Wait()
}

func verifyBlkTxVersion(blkVer uint8, txVer uint8) bool {
	if blkVer == Genesis {
		return txVer == Genesis
	}

	return txVer != Genesis
}
