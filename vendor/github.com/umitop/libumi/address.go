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
	"encoding/binary"
)

// AddressLength ...
const AddressLength = 34

const (
	genesis uint16 = 0
	umi     uint16 = 21929
)

// Address ...
type Address []byte

// NewAddressFromBech32 ...
func NewAddressFromBech32(s string) (adr Address, err error) {
	pfx, pub, err := bech32Decode(s)
	if err != nil {
		return adr, err
	}

	adr = NewAddressBuilder().SetPrefix(pfx).SetPublicKey(pub).Build()

	return adr, err
}

// Bech32 ...
func (a Address) Bech32() string {
	return bech32Encode(a.Prefix(), a.PublicKey())
}

// Version ...
func (a Address) Version() uint16 {
	return binary.BigEndian.Uint16(a[0:2])
}

// Prefix ...
func (a Address) Prefix() string {
	return versionToPrefix(binary.BigEndian.Uint16(a[0:2]))
}

// PublicKey ...
func (a Address) PublicKey() []byte {
	return a[2:34]
}

// Verify ...
func (a Address) Verify() error {
	return assert(a,
		lengthIs(AddressLength),
		versionIsValid,
	)
}
