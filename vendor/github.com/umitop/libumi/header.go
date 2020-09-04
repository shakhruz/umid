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

// Header ...
type Header []byte

// Hash ...
func (h Header) Hash() []byte {
	s := sha256.Sum256(h)

	return s[:]
}

// MerkleRootHash ...
func (h Header) MerkleRootHash() []byte {
	return h[33:65]
}

// PreviousBlockHash ...
func (h Header) PreviousBlockHash() []byte {
	return h[1:33]
}

// PublicKey ...
func (h Header) PublicKey() ed25519.PublicKey {
	return ed25519.PublicKey(h[71:103])
}

// Signature ...
func (h Header) Signature() []byte {
	return h[103:167]
}

// Timestamp ..
func (h Header) Timestamp() uint32 {
	return binary.BigEndian.Uint32(h[65:69])
}

// TxCount ...
func (h Header) TxCount() uint16 {
	return binary.BigEndian.Uint16(h[69:71])
}

// Verify ...
func (h Header) Verify() (ok bool, err error) {
	ok = ed25519.Verify(h.PublicKey(), h[0:103], h.Signature())
	if !ok {
		err = ErrInvalidSignature
	}

	return ok, err
}

// Version ...
func (h Header) Version() uint8 {
	return h[0]
}
