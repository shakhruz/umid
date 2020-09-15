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

import "strings"

const (
	prefixAbc  = " abcdefghijklmnopqrstuvwxyz"
	bech32Abc  = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	pfxGenesis = "genesis"
	pfxLen     = 3
	dataLen    = 59
	addressLen = 62
)

func bech32Encode(pfx string, pub []byte) string {
	data := bech32Convert8to5(pub)

	var s strings.Builder

	s.Grow(addressLen)
	s.WriteString(pfx)
	s.WriteString("1")
	s.Write(data)
	s.Write(bech32CreateChecksum(pfx, data))

	return s.String()
}

func bech32Decode(bech string) (pfx string, data []byte, err error) {
	bech = strings.ToLower(bech)

	if pfx, err = bech32ParsePrefix(bech); err != nil {
		return
	}

	if data, err = bech32Convert5to8([]byte(bech[len(pfx)+1 : len(bech)-6])); err != nil {
		return
	}

	if !bech32VerifyChecksum(pfx, []byte(bech[len(pfx)+1:])) {
		err = ErrInvalidAddress
	}

	return pfx, data, err
}

func bech32ParsePrefix(s string) (string, error) {
	sep := strings.LastIndexByte(s, '1')
	if sep == -1 {
		return "", ErrInvalidAddress
	}

	if len(s)-sep != dataLen {
		return "", ErrInvalidAddress
	}

	pfx := s[0:sep]

	if !bech32VerifyPrefix(pfx) {
		return "", ErrInvalidAddress
	}

	return pfx, nil
}

func bech32VerifyPrefix(s string) bool {
	if s == pfxGenesis {
		return true
	}

	if len(s) != pfxLen {
		return false
	}

	for i := range s {
		if strings.IndexByte(prefixAbc, s[i]) < 1 {
			return false
		}
	}

	return true
}

func bech32Convert5to8(data []byte) ([]byte, error) {
	var acc, bits int

	out := make([]byte, 0, 32)

	for _, b := range data {
		v := strings.IndexByte(bech32Abc, b)
		if v == -1 {
			return nil, ErrInvalidAddress
		}

		acc = (acc << 5) | v
		bits += 5

		for bits >= 8 {
			bits -= 8
			out = append(out, byte(acc>>bits))
		}
	}

	return out, bech32Convert5to8Verify(acc, bits)
}

func bech32Convert5to8Verify(acc, bits int) error {
	if bits >= 5 || (acc<<(8-bits))&0xff > 0 {
		return ErrInvalidAddress
	}

	return nil
}

func bech32Convert8to5(data []byte) []byte {
	var acc, bits int

	res := make([]byte, 0, 52)

	for _, b := range data {
		acc = (acc << 8) | int(b)
		bits += 8

		for bits >= 5 {
			bits -= 5
			res = append(res, bech32Abc[acc>>bits&0x1f])
		}
	}

	if bits > 0 {
		res = append(res, bech32Abc[acc<<(5-bits)&0x1f])
	}

	return res
}

func bech32CreateChecksum(prefix string, data []byte) []byte {
	values := bech32PrefixExpand(prefix)

	for _, v := range data {
		values = append(values, strings.IndexByte(bech32Abc, v))
	}

	polymod := bech32PolyMod(append(values, 0, 0, 0, 0, 0, 0)) ^ 1
	checksum := make([]byte, 6)

	for i := range checksum {
		checksum[i] = bech32Abc[byte((polymod>>uint(5*(5-i)))&31)] //nolint:gomnd
	}

	return checksum
}

func bech32PolyMod(values []int) int {
	chk := 1

	for _, v := range values {
		b := chk >> 25               //nolint:gomnd
		chk = (chk&0x1ffffff)<<5 ^ v //nolint:gomnd
		chk = bech32PolyModGen(chk, b)
	}

	return chk
}

func bech32PolyModGen(chk, b int) int {
	for i, g := range [...]int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3} {
		if (b>>i)&1 == 1 {
			chk ^= g
		}
	}

	return chk
}

func bech32PrefixExpand(p string) []int {
	l := len(p)
	r := make([]int, l*2+1)

	for i, s := range p {
		r[i] = int(s) >> 5     //nolint:gomnd
		r[i+l+1] = int(s) & 31 //nolint:gomnd
	}

	return r
}

func bech32VerifyChecksum(prefix string, data []byte) bool {
	b := bech32PrefixExpand(prefix)

	for _, v := range data {
		b = append(b, strings.IndexByte(bech32Abc, v))
	}

	return bech32PolyMod(b) == 1
}
