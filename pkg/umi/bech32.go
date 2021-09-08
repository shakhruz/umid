// Copyright (c) 2021 UMI
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

package umi

import (
	"bytes"
	"strings"
)

// @see https://medium.com/@MeshCollider/some-of-the-math-behind-bech32-addresses-cf03c7496285
// @see https://github.com/bitcoin/bips/blob/master/bip-0173.mediawiki

const abc = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func Bech32Encode(hrp string, data []byte) (bech32 string) {
	var s strings.Builder

	// [3]hrp + '1' + [52]data + [6]checksum
	s.Grow(62)

	// hrp
	_, _ = s.WriteString(hrp)

	// separator
	_ = s.WriteByte('1')

	// data
	pub := bits8to5(data)
	for _, i := range pub {
		_ = s.WriteByte(abc[i])
	}

	// checksum
	chk := checksum(hrp, pub)
	for _, i := range chk {
		_ = s.WriteByte(abc[i])
	}

	return s.String()
}

func Bech32Decode(bech32 string) (hrp string, data []byte) {
	data = make([]byte, 52)

	switch len(bech32) {
	case 62:
		hrp = bech32[0:3]
		copy(data[0:52], bech32[4:56])
	case 66:
		hrp = bech32[0:7]
		copy(data[0:52], bech32[8:60])
	}

	for i := 0; i < 52; i++ {
		data[i] = byte(strings.IndexByte(abc, data[i]))
	}

	data = bits5to8(data)

	return hrp, data
}

func IsBech32Valid(bech32 string) bool {
	var hrp string

	data := make([]byte, 58)

	switch {
	case len(bech32) == 62 && bech32[3] == '1': // адрес имеет длину 62 символа, префикс 3 символа.
		hrp = bech32[:3]
		copy(data, bech32[4:])
	case len(bech32) == 66 && strings.HasPrefix(bech32, "genesis1"): // Genesis-адрес имеет длину 66 символов.
		hrp = bech32[:7]
		copy(data, bech32[8:])
	default:
		return false
	}

	for i := 0; i < 58; i++ {
		data[i] = byte(strings.IndexByte(abc, data[i]))
	}

	return bytes.Equal(data[52:], checksum(hrp, data[:52]))
}

func bits8to5(b8 []byte) (b5 []byte) {
	var (
		bits uint8
		val  uint16
	)

	b5 = make([]byte, 0, 52)

	for _, b := range b8 {
		val = (val << 8) | uint16(b)
		bits += 8

		for bits >= 5 {
			bits -= 5
			b5 = append(b5, byte(val>>bits&31))
		}
	}

	b5 = append(b5, byte(val<<(5-bits)&31))

	return b5
}

func bits5to8(b5 []byte) (b8 []byte) {
	var (
		bits uint8
		val  uint16
	)

	b8 = make([]byte, 0, 32)

	for _, b := range b5 {
		val = (val << 5) | uint16(b)
		bits += 5

		for bits >= 8 {
			bits -= 8
			b8 = append(b8, byte(val>>bits&255))
		}
	}

	return b8
}

func polyMod(data []byte) (mod int32) {
	mod = 1

	for _, v := range data {
		b := mod >> 25
		mod = (mod&0x01ffffff)<<5 ^ int32(v)

		for i, g := range [5]int32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3} {
			if (b>>i)&1 == 1 {
				mod ^= g
			}
		}
	}

	return mod
}

func checksum(hrp string, data []byte) []byte {
	// [high bits of HRP] + [0] + [low bits of HRP] + [data] + [0,0,0,0,0,0]
	val := make([]byte, 0, 65)

	// high bits of HRP
	for _, b := range hrp {
		val = append(val, byte(b)>>5)
	}

	// [0]
	val = append(val, 0)

	// low bits of HRP
	for _, b := range hrp {
		val = append(val, byte(b)&31)
	}

	// data
	val = append(val, data...)

	// [0,0,0,0,0,0]
	val = append(val, 0, 0, 0, 0, 0, 0)

	// PolyMod returns an integer, and then that integer is XOR’ed with 1.
	mod := polyMod(val) ^ 1

	// The resulting integer is then split up into six 5-bit blocks just like the rest
	// of the data list (ret[i] = (mod >> (5 * (5-i))) & 31), so that each 5 bits
	// can be encoded in a Base32 character as well.
	sum := make([]byte, 6)
	sum[0] = byte(mod >> 25 & 31)
	sum[1] = byte(mod >> 20 & 31)
	sum[2] = byte(mod >> 15 & 31)
	sum[3] = byte(mod >> 10 & 31)
	sum[4] = byte(mod >> 5 & 31)
	sum[5] = byte(mod & 31)

	return sum
}
