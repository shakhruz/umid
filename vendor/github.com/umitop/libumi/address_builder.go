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
	"strings"
)

// AddressBuilder ...
type AddressBuilder []byte

// NewAddressBuilder ...
func NewAddressBuilder() AddressBuilder {
	adr := make(AddressBuilder, AddressLength)
	adr.SetVersion(umi)

	return adr
}

// SetVersion ...
func (a AddressBuilder) SetVersion(v uint16) AddressBuilder {
	binary.BigEndian.PutUint16(a, v)

	return a
}

// SetPrefix ...
func (a AddressBuilder) SetPrefix(s string) AddressBuilder {
	binary.BigEndian.PutUint16(a[0:2], prefixToVersion(s))

	return a
}

// SetPublicKey ...
func (a AddressBuilder) SetPublicKey(b []byte) AddressBuilder {
	copy(a[2:34], b)

	return a
}

// Build ...
func (a AddressBuilder) Build() Address {
	adr := make(Address, AddressLength)
	copy(adr, a)

	return adr
}

func prefixToVersion(s string) (v uint16) {
	if s != pfxGenesis {
		for i := range s {
			v |= setChrToVer(s[i], i)
		}
	}

	return v
}

func setChrToVer(c byte, i int) uint16 {
	return (uint16(c) - 96) << ((2 - i) * 5)
}

func versionToPrefix(v uint16) string {
	if v == genesis {
		return pfxGenesis
	}

	var s strings.Builder

	s.Grow(pfxLen)
	s.WriteByte(getChrFromVer(v, 0))
	s.WriteByte(getChrFromVer(v, 1))
	s.WriteByte(getChrFromVer(v, 2))

	return s.String()
}

func getChrFromVer(v uint16, i int) byte {
	return byte(v>>((2-i)*5))&31 + 96
}
