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

const (
	PfxLength = 2
)

const (
	PfxVerGenesis Prefix = 0x0000
	PfxVerUmi     Prefix = 0x55A9
	PfxVerRoy     Prefix = 0x49F9
	PfxVerIsp     Prefix = 0x2670
	PfxVerGls     Prefix = 0x1d93
	PfxVerGlz     Prefix = 0x1d9a
	PfxVerNft     Prefix = 0x38d4

	pfxGenesis string = "genesis"
)

type Prefix uint16

func (prefix Prefix) Bytes() []byte {
	if prefix == PfxVerGenesis {
		return []byte(pfxGenesis)
	}

	buf := make([]byte, 3)
	buf[0] = byte((prefix>>10)&31) + 96
	buf[1] = byte((prefix>>5)&31) + 96
	buf[2] = byte(prefix&31) + 96

	return buf
}

func (prefix Prefix) String() string {
	if prefix == PfxVerGenesis {
		return pfxGenesis
	}

	return string(prefix.Bytes())
}

func (prefix Prefix) IsValid() bool {
	switch prefix {
	case PfxVerGenesis, PfxVerUmi, PfxVerRoy, PfxVerIsp:
		return true
	}

	hrp := prefix.String()

	return (uint16)(prefix)>>15 == 0 && // первый бит должен быть равен нулю
		hrp[0] >= 'a' && hrp[0] <= 'z' &&
		hrp[1] >= 'a' && hrp[1] <= 'z' &&
		hrp[2] >= 'a' && hrp[2] <= 'z'
}

func ParsePrefix(hrp string) Prefix {
	if hrp == pfxGenesis {
		return PfxVerGenesis
	}

	chr0 := uint16(hrp[0]-96) & 31
	chr1 := uint16(hrp[1]-96) & 31
	chr2 := uint16(hrp[2]-96) & 31

	return (Prefix)((chr0 << 10) | (chr1 << 5) | chr2)
}

func VerifyHrp(hrp string) bool {
	return len(hrp) == 3 &&
		hrp[0] >= 'a' && hrp[0] <= 'z' &&
		hrp[1] >= 'a' && hrp[1] <= 'z' &&
		hrp[2] >= 'a' && hrp[2] <= 'z'
}
