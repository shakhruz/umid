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
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	AddrLength = 34
)

type Address [AddrLength]byte

var ErrAddress = errors.New("address")

func ParseAddress(bech32 string) (Address, error) {
	var address Address

	if !IsBech32Valid(bech32) {
		return address, fmt.Errorf("%w: malformed bech32 address", ErrAddress)
	}

	hrp, pubicKey := Bech32Decode(bech32)
	prefix := ParsePrefix(hrp)

	if !prefix.IsValid() {
		return address, fmt.Errorf("%w: malformed bech32 address prefix", ErrAddress)
	}

	address.SetPrefix(prefix)
	address.SetPublicKey(pubicKey)

	return address, nil
}

func (address Address) String() string {
	return Bech32Encode(address.Prefix().String(), address.PublicKey())
}

func (address Address) Prefix() Prefix {
	return Prefix(binary.BigEndian.Uint16(address[:PfxLength]))
}

func (address *Address) SetPrefix(prefix Prefix) {
	binary.BigEndian.PutUint16(address[:PfxLength], uint16(prefix))
}

func (address Address) PublicKey() PublicKey {
	return address[PfxLength:]
}

func (address *Address) SetPublicKey(publicKey PublicKey) {
	copy(address[PfxLength:], publicKey)
}
