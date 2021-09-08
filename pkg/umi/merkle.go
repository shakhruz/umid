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
	"crypto/sha256"
	"errors"
)

var ErrUniq = errors.New("block contains non uniq transaction")

func MerkleRoot(txs []byte) [32]byte {
	cur := len(txs) / TxLength

	if cur == 1 {
		return sha256.Sum256(txs)
	}

	nxt := (cur + cur&1) / 2
	hsh := make([][32]byte, nxt)
	tmp := make([]byte, 64)

	for i := 0; i < nxt; i++ {
		l := i * 300
		h := l + 150

		x := sha256.Sum256(txs[l:h])
		copy(tmp[:32], x[:])

		if h < len(txs) {
			x = sha256.Sum256(txs[l+150 : h+150])
		}

		copy(tmp[32:], x[:])
		hsh[i] = sha256.Sum256(tmp)
	}

	for {
		if nxt == 1 {
			break
		}

		cur = nxt
		nxt = (cur + cur&1) / 2

		for i := 0; i < nxt; i++ {
			j := i * 2
			copy(tmp[:32], hsh[j][:])

			if j < cur-1 {
				j++
			}

			copy(tmp[32:], hsh[j][:])
			hsh[i] = sha256.Sum256(tmp)
		}
	}

	return hsh[0]
}

func MerkleRootUniq(txs []byte) ([]byte, error) {
	res := make([]byte, 32)
	cur := len(txs) / TxLength

	if cur == 1 {
		x := sha256.Sum256(txs)
		copy(res, x[:])

		return res, nil
	}

	unq := make(map[[32]byte]struct{}, cur)
	hsh := make([][32]byte, cur)

	for i := 0; i < cur; i++ {
		l := i * 150
		h := l + 150

		x := sha256.Sum256(txs[l:h])

		if _, ok := unq[x]; ok {
			return nil, ErrUniq
		}

		unq[x] = struct{}{}
		hsh[i] = x
	}

	tmp := make([]byte, 64)

	for {
		if cur == 1 {
			break
		}

		nxt := (cur + cur&1) / 2

		for i := 0; i < nxt; i++ {
			j := i * 2
			copy(tmp[:32], hsh[j][:])

			if j < cur-1 {
				j++
			}

			copy(tmp[32:], hsh[j][:])
			hsh[i] = sha256.Sum256(tmp)
		}

		cur = nxt
	}

	copy(res, hsh[0][:])

	return res, nil
}
