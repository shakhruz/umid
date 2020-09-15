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

package blockchain

import (
	"encoding/hex"
	"errors"
	"log"

	"github.com/umitop/libumi"
)

var errInvalidPubKey = errors.New("block signed by unapproved key")

type iBlock interface {
	AddBlock(raw []byte) error
	GetLastBlockHeight() (height uint64, err error)
	GetLastBlockHash() (hash []byte, err error)
	ListBlocksAfterKey(key uint64, lim uint16) (raws [][]byte, err error)
}

// AddBlock ...
func (b *block) AddBlock(raw []byte) error {
	if err := b.VerifyBlock(raw); err != nil {
		return err
	}

	return b.db.AddBlock(raw)
}

// LastBlockHeight ...
func (b *block) GetLastBlockHeight() (height uint64, err error) {
	return b.db.GetLastBlockHeight()
}

// GetLastBlockHash ...
func (b *block) GetLastBlockHash() (hash []byte, err error) {
	return b.db.GetLastBlockHash()
}

// BlocksByHeight ...
func (b *block) ListBlocksAfterKey(key uint64, lim uint16) (raws [][]byte, err error) {
	return b.db.ListBlocksAfterKey(key, lim)
}

// VerifyBlock ...
func (b *block) VerifyBlock(raw []byte) error {
	blk := (libumi.Block)(raw)

	if len(blk) < libumi.HeaderLength {
		return errInvalidLength
	}

	if _, ok := b.approvedKeys[string(blk.PublicKey())]; !ok {
		log.Printf("block %X has invalid public key\n", (libumi.Block)(raw).Hash())

		return errInvalidPubKey
	}

	return libumi.VerifyBlock(raw)
}

// ValidateBlock ...
func (b *block) ValidateBlock(raw []byte) error {
	blk := (libumi.Block)(raw)
	l := blk.TxCount()

	for i := uint16(0); i < l; i++ {
		_ = blk.Transaction(i)
	}

	return nil
}

type block struct {
	db           iBlock
	approvedKeys map[string]struct{}
}

func newBlock(db iBlock) (b block) {
	keys := []string{
		"45885c9687d799a4d1f4d786d8639274d293ed024ad7f7a436715d6217c9f72b",
	}

	keyz := make(map[string]struct{}, len(keys))

	for _, key := range keys {
		k, _ := hex.DecodeString(key)
		keyz[string(k)] = struct{}{}
	}

	return block{
		db:           db,
		approvedKeys: keyz,
	}
}
