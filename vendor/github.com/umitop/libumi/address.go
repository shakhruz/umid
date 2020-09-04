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
	"errors"
	"strings"
)

const (
	// AddressLength ...
	AddressLength = 34

	alphabet = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
)

var (
	// ErrAddrInvalidPrefix ...
	ErrAddrInvalidPrefix = errors.New("address: invalid prefix")
	// ErrBechInvalidCharacter ...
	ErrBechInvalidCharacter = errors.New("bech32: invalid character")
	// ErrBechInvalidChecksum ...
	ErrBechInvalidChecksum = errors.New("bech32: invalid checksum")
	// ErrBechInvalidData ...
	ErrBechInvalidData = errors.New("bech32: invalid data")
	// ErrBechInvalidLength ...
	ErrBechInvalidLength = errors.New("bech32: invalid length")
	// ErrBechMissingSeparator ...
	ErrBechMissingSeparator = errors.New("bech32: missing separator")
)

// Address ...
type Address []byte

// NewAddressFromBech32 ...
func NewAddressFromBech32(s string) (adr Address, err error) {
	prefix, bytes, err := decode(s)
	if err != nil {
		return adr, err
	}

	adr = NewAddress()
	copy(adr[2:], bytes)
	err = adr.SetPrefix(prefix)

	return adr, err
}

// NewAddressFromPublicKey ...
func NewAddressFromPublicKey(k []byte) Address {
	adr := NewAddress()
	copy(adr[2:34], k)

	return adr
}

// NewAddress ...
func NewAddress() Address {
	adr := make(Address, AddressLength)
	_ = adr.SetPrefix("umi")

	return adr
}

// NewAddressWithPrefix ...
func NewAddressWithPrefix(p string) Address {
	adr := make(Address, AddressLength)
	_ = adr.SetPrefix(p)

	return adr
}

// Bech32 ...
func (a Address) Bech32() string {
	return encode(a.Prefix(), a.PublicKey())
}

// Prefix ...
func (a Address) Prefix() string {
	if a[0] == 0 && a[1] == 0 {
		return "genesis"
	}

	var p [3]byte
	p[0] = ((a[0] >> 2) & 31) + 96
	p[1] = (((a[0] & 3) << 3) | (a[1] >> 5)) + 96
	p[2] = (a[1] & 31) + 96

	return string(p[:])
}

// SetPrefix ...
func (a Address) SetPrefix(p string) (err error) {
	if p == "genesis" {
		a[0], a[1] = 0, 0

		return err
	}

	if len(p) != 3 {
		return ErrAddrInvalidPrefix
	}

	for _, c := range p {
		if c < 97 || c > 122 {
			return ErrAddrInvalidPrefix
		}
	}

	a[0] = (((p[0] - 96) & 31) << 2) | (((p[1] - 96) & 31) >> 3)
	a[1] = ((p[1] - 96) << 5) | ((p[2] - 96) & 31)

	return err
}

// PublicKey ...
func (a Address) PublicKey() []byte {
	return a[2:34]
}

func encode(prefix string, bytes []byte) string {
	data := convert8to5(bytes)

	var s strings.Builder

	s.Grow(59 + len(prefix))
	s.WriteString(prefix)
	s.WriteString("1")
	s.Write(data)
	s.Write(createChecksum(prefix, data))

	return s.String()
}

func decode(bech string) (pfx string, bytes []byte, err error) {
	if len(bech) != 62 && len(bech) != 66 {
		return pfx, bytes, ErrBechInvalidLength
	}

	bech = strings.ToLower(bech)

	sep := strings.LastIndexByte(bech, '1')
	if sep == -1 {
		return pfx, bytes, ErrBechMissingSeparator
	}

	bytes, err = convert5to8([]byte(bech[sep+1 : len(bech)-6]))
	if err != nil {
		return pfx, nil, err
	}

	pfx = bech[0:sep]

	if !verifyChecksum(pfx, []byte(bech[sep+1:])) {
		return pfx, nil, ErrBechInvalidChecksum
	}

	return pfx, bytes[0:32], err
}

func convert5to8(data []byte) (out []byte, err error) {
	var acc, bits int

	out = make([]byte, 0, 32)

	for _, b := range data {
		v := strings.IndexByte(alphabet, b)
		if v == -1 {
			return nil, ErrBechInvalidCharacter
		}

		acc = (acc << 5) | v
		bits += 5

		for bits >= 8 {
			bits -= 8
			out = append(out, byte(acc>>bits&0xff))
		}
	}

	if bits >= 5 || (acc<<(8-bits))&0xff > 0 {
		return nil, ErrBechInvalidData
	}

	return out, err
}

func convert8to5(data []byte) []byte {
	var acc, bits int

	res := make([]byte, 0, 52)

	for _, b := range data {
		acc = (acc << 8) | int(b)
		bits += 8

		for bits >= 5 {
			bits -= 5
			res = append(res, alphabet[acc>>bits&0x1f])
		}
	}

	if bits > 0 {
		res = append(res, alphabet[acc<<(5-bits)&0x1f])
	}

	return res
}

func createChecksum(prefix string, data []byte) []byte {
	b := prefixExpand(prefix)

	for _, v := range data {
		b = append(b, strings.IndexByte(alphabet, v))
	}

	b = append(b, 0, 0, 0, 0, 0, 0)
	p := polyMod(b) ^ 1

	c := make([]byte, 6)
	for i := 0; i < 6; i++ {
		c[i] = alphabet[byte((p>>uint(5*(5-i)))&31)]
	}

	return c
}

func polyMod(values []int) int {
	generator := []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := 1

	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v

		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= generator[i]
			}
		}
	}

	return chk
}

func prefixExpand(p string) []int {
	l := len(p)
	r := make([]int, l*2+1, l*2+59)

	for i := 0; i < l; i++ {
		r[i] = int(p[i]) >> 5
		r[i+l+1] = int(p[i]) & 31
	}

	return r
}

func verifyChecksum(prefix string, data []byte) bool {
	b := prefixExpand(prefix)

	for _, v := range data {
		b = append(b, strings.IndexByte(alphabet, v))
	}

	return polyMod(b) == 1
}
