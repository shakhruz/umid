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
	"errors"

	"github.com/umitop/libumi"
)

var errInvalidPubKey = errors.New("invalid key")

// AddBlock ...
func (bc *Blockchain) AddBlock(b []byte) error {
	if err := bc.VerifyBlock(b); err != nil {
		return err
	}

	return bc.storage.AddBlock(b)
}

// LastBlockHeight ...
func (bc *Blockchain) LastBlockHeight() (uint32, error) {
	return bc.storage.LastBlockHeight()
}

// VerifyBlock ...
func (bc *Blockchain) VerifyBlock(b []byte) error {
	blk := (libumi.Block)(b)

	if _, ok := bc.approvedKeys[string(blk.PublicKey())]; !ok {
		return errInvalidPubKey
	}

	return blk.Verify()
}

// ValidateBlock ...
func (bc *Blockchain) ValidateBlock(b []byte) error {
	blk := (libumi.Block)(b)
	l := blk.TxCount()

	for i := uint16(0); i < l; i++ {
		_ = blk.Transaction(i)
	}

	return nil
}

// BlocksByHeight ...
func (bc *Blockchain) BlocksByHeight(n uint64) ([][]byte, error) {
	return bc.storage.BlocksByHeight(n)
}
