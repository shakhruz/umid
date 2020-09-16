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
	"encoding/binary"
	"time"
)

// TransactionBuilder ...
type TransactionBuilder []byte

// NewTransactionBuilder ...
func NewTransactionBuilder() TransactionBuilder {
	tx := make(TransactionBuilder, TxLength)
	tx.SetVersion(Basic)

	return tx
}

// SetVersion ...
func (t TransactionBuilder) SetVersion(n uint8) TransactionBuilder {
	t[0] = n

	return t
}

// SetSender ...
func (t TransactionBuilder) SetSender(a Address) TransactionBuilder {
	copy(t[1:35], a)

	return t
}

// SetRecipient ...
func (t TransactionBuilder) SetRecipient(a Address) TransactionBuilder {
	copy(t[35:69], a)

	return t
}

// SetValue ...
func (t TransactionBuilder) SetValue(n uint64) TransactionBuilder {
	binary.BigEndian.PutUint64(t[69:77], n)

	return t
}

// SetPrefix ...
func (t TransactionBuilder) SetPrefix(s string) TransactionBuilder {
	binary.BigEndian.PutUint16(t[35:37], prefixToVersion(s))

	return t
}

// SetProfitPercent ...
func (t TransactionBuilder) SetProfitPercent(n uint16) TransactionBuilder {
	binary.BigEndian.PutUint16(t[37:39], n)

	return t
}

// SetFeePercent ...
func (t TransactionBuilder) SetFeePercent(p uint16) TransactionBuilder {
	binary.BigEndian.PutUint16(t[39:41], p)

	return t
}

// SetName ...
func (t TransactionBuilder) SetName(s string) TransactionBuilder {
	t[41] = uint8(len(s))
	copy(t[42:77], s)

	return t
}

// SetNonce ...
func (t TransactionBuilder) SetNonce(n uint64) TransactionBuilder {
	binary.BigEndian.PutUint64(t[77:85], n)

	return t
}

// SetSignature ...
func (t TransactionBuilder) SetSignature(sig []byte) TransactionBuilder {
	copy(t[85:149], sig)

	return t
}

// Build ...
func (t TransactionBuilder) Build() Transaction {
	tx := make(Transaction, TxLength)
	copy(tx, t)

	return tx
}

// Sign ...
func (t TransactionBuilder) Sign(sec []byte) TransactionBuilder {
	t.SetNonce(uint64(time.Now().UnixNano()))
	t.SetSignature(ed25519.Sign(sec, t[0:85]))

	return t
}
