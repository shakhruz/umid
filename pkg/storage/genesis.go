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

package storage

import (
	"gitlab.com/umitop/umid/pkg/umi"
)

func GenesisBlock(network string) (block umi.Block) {
	switch network {
	case "testnet":
		block = []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6e, 0xa1, 0x8c,
			0xf7, 0x13, 0x7e, 0x30, 0xc1, 0xc1, 0xca, 0xfd, 0x64, 0x16, 0xde, 0xbd, 0x7a, 0x63, 0xf8, 0x47, 0xa0, 0xaa,
			0xbf, 0x04, 0x20, 0xd2, 0xd2, 0x9e, 0xf3, 0x6c, 0x0c, 0xbd, 0xa9, 0x5e, 0xab, 0x66, 0x80, 0x00, 0x01, 0x46,
			0x52, 0xd6, 0x09, 0x7f, 0x24, 0x43, 0x4f, 0xac, 0xbc, 0x18, 0x28, 0x94, 0xfa, 0xfa, 0x4f, 0xc5, 0xad, 0xe7,
			0x8d, 0xb3, 0xe0, 0x17, 0x05, 0xa6, 0x32, 0x4d, 0x35, 0xd8, 0xd3, 0x4a, 0x69, 0x02, 0xdf, 0x84, 0xb6, 0xf5,
			0x1f, 0xef, 0xfa, 0x62, 0x06, 0xbf, 0x68, 0x38, 0x7f, 0x34, 0xd0, 0xf6, 0x0f, 0x61, 0x43, 0x88, 0x55, 0x76,
			0xe6, 0x42, 0xa8, 0x53, 0x6d, 0xa3, 0xc7, 0x53, 0xc6, 0x84, 0x1b, 0x43, 0x23, 0x21, 0x03, 0x4c, 0xe9, 0x02,
			0xf4, 0xc4, 0x16, 0xc8, 0x32, 0x68, 0x9e, 0x24, 0x8d, 0x77, 0x70, 0x68, 0x34, 0xc4, 0xf7, 0x48, 0x2d, 0x30,
			0x83, 0x42, 0xca, 0x32, 0x08, 0x00, 0x00, 0x00, 0x46, 0x52, 0xd6, 0x09, 0x7f, 0x24, 0x43, 0x4f, 0xac, 0xbc,
			0x18, 0x28, 0x94, 0xfa, 0xfa, 0x4f, 0xc5, 0xad, 0xe7, 0x8d, 0xb3, 0xe0, 0x17, 0x05, 0xa6, 0x32, 0x4d, 0x35,
			0xd8, 0xd3, 0x4a, 0x69, 0x55, 0xa9, 0x5c, 0xa1, 0x96, 0x34, 0xe7, 0xaa, 0x04, 0xbf, 0x1f, 0x13, 0x56, 0x89,
			0x2d, 0xdb, 0x31, 0x59, 0x32, 0x3a, 0xe5, 0xeb, 0x49, 0xbf, 0xea, 0x90, 0x07, 0xe4, 0xc6, 0xd2, 0xc1, 0xee,
			0x5c, 0x2a, 0x00, 0x00, 0x00, 0x00, 0x6b, 0x49, 0xd2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x4a, 0x50, 0x0d, 0x07, 0x27, 0x11, 0x70, 0x63, 0x18, 0x71, 0xec, 0xc6, 0x60, 0xf1, 0xb3, 0xf5, 0x5d, 0x2e,
			0x36, 0x77, 0x2b, 0x53, 0x2c, 0xb2, 0x64, 0x0e, 0x10, 0x59, 0xa3, 0x99, 0x6f, 0x87, 0x71, 0x8b, 0x97, 0x56,
			0x38, 0x72, 0x0c, 0xe2, 0x75, 0x38, 0xa8, 0x1b, 0xb0, 0x4f, 0x7f, 0x1c, 0x92, 0xe4, 0x3d, 0x7f, 0x25, 0x47,
			0x4a, 0x39, 0xb2, 0x4e, 0x19, 0x9c, 0xc8, 0x29, 0x53, 0x0f, 0x00, 0x5e, 0xab, 0x66, 0x80, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x6b,
			0x49, 0xd2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00,
		}
	default: // case "mainnet":
		block = []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf1, 0x21, 0x5c,
			0xf6, 0x9c, 0x9a, 0x4a, 0xf0, 0xfc, 0xde, 0x82, 0xa8, 0x5c, 0xb2, 0x0a, 0x32, 0x23, 0xc4, 0xe1, 0x98, 0xfe,
			0xdf, 0x25, 0x58, 0x2c, 0x44, 0x77, 0x30, 0x34, 0x85, 0x35, 0x17, 0x5e, 0xd4, 0x45, 0x00, 0x00, 0x01, 0x45,
			0x88, 0x5c, 0x96, 0x87, 0xd7, 0x99, 0xa4, 0xd1, 0xf4, 0xd7, 0x86, 0xd8, 0x63, 0x92, 0x74, 0xd2, 0x93, 0xed,
			0x02, 0x4a, 0xd7, 0xf7, 0xa4, 0x36, 0x71, 0x5d, 0x62, 0x17, 0xc9, 0xf7, 0x2b, 0x1c, 0xe8, 0xd4, 0x01, 0xc6,
			0xc2, 0x10, 0x61, 0xc9, 0x14, 0x5f, 0x94, 0xc5, 0x2c, 0x9a, 0xa9, 0x1a, 0x64, 0x5d, 0xa7, 0x2d, 0x51, 0xc0,
			0x2b, 0x11, 0xf1, 0x2e, 0x5c, 0x8d, 0xa6, 0xfd, 0xe6, 0x76, 0x2b, 0x29, 0x06, 0xef, 0x77, 0x97, 0xa9, 0x45,
			0x5a, 0x3b, 0xaa, 0xa0, 0xff, 0x86, 0x99, 0xc6, 0x64, 0x3f, 0x98, 0xfb, 0xd3, 0xc7, 0x38, 0x1c, 0x2c, 0x5e,
			0xfa, 0x64, 0x18, 0x94, 0x03, 0x00, 0x00, 0x00, 0x45, 0x88, 0x5c, 0x96, 0x87, 0xd7, 0x99, 0xa4, 0xd1, 0xf4,
			0xd7, 0x86, 0xd8, 0x63, 0x92, 0x74, 0xd2, 0x93, 0xed, 0x02, 0x4a, 0xd7, 0xf7, 0xa4, 0x36, 0x71, 0x5d, 0x62,
			0x17, 0xc9, 0xf7, 0x2b, 0x55, 0xa9, 0x4f, 0x26, 0x03, 0xe8, 0xad, 0xcb, 0xf4, 0x63, 0xad, 0xa9, 0x5d, 0xdf,
			0x78, 0x50, 0x47, 0x64, 0x7c, 0x3e, 0xfe, 0x24, 0x27, 0x7f, 0xbe, 0x32, 0x5e, 0x93, 0xc4, 0x5a, 0x8d, 0x66,
			0xc7, 0x4c, 0x00, 0x00, 0x00, 0x00, 0x6b, 0x49, 0xd2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0xf2, 0x45, 0x52, 0xd3, 0xd6, 0xd6, 0x19, 0xb3, 0x72, 0x0c, 0x09, 0x56, 0x31, 0xa7, 0xe6, 0xc5, 0xe8, 0xaf,
			0xf4, 0x8f, 0xd5, 0xdd, 0xa0, 0x23, 0x79, 0xe2, 0xa0, 0x2f, 0x5c, 0x7b, 0x02, 0x23, 0x5c, 0x7c, 0x09, 0xf1,
			0x77, 0xd2, 0x87, 0x09, 0x97, 0xe1, 0xbd, 0x4b, 0x0e, 0x96, 0xd6, 0x88, 0x16, 0x78, 0xea, 0x18, 0x58, 0xc3,
			0x6f, 0x6c, 0x6d, 0x35, 0x6b, 0xf6, 0x7b, 0xa9, 0x5d, 0x02, 0x00, 0x5e, 0xd4, 0x45, 0x00, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x6b,
			0x49, 0xd2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00,
		}
	}

	return block
}
