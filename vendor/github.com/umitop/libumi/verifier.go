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
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"runtime"
	"sync"
	"unicode/utf8"
)

// Errors.
var (
	ErrInvalidLength        = errors.New("invalid length")
	ErrInvalidVersion       = errors.New("invalid version")
	ErrInvalidSender        = errors.New("invalid sender")
	ErrInvalidRecipient     = errors.New("invalid recipient")
	ErrInvalidPrefix        = errors.New("invalid prefix")
	ErrInvalidProfitPercent = errors.New("invalid profit percent")
	ErrInvalidFeePercent    = errors.New("invalid fee percent")
	ErrInvalidName          = errors.New("invalid name")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrInvalidPrevHash      = errors.New("invalid previous block hash")
	ErrInvalidMerkle        = errors.New("invalid merkle root")
	ErrInvalidTx            = errors.New("invalid transaction")
	ErrNonUniqueTx          = errors.New("non-unique transaction")
)

// ErrInvalidAddress ...
var ErrInvalidAddress = errors.New("invalid address")

func assert(b []byte, asserts ...func([]byte) error) error {
	return runAsserts(b, asserts)
}

func runAsserts(b []byte, asserts []func([]byte) error) error {
	for _, assert := range asserts {
		if err := assert(b); err != nil {
			return err
		}
	}

	return nil
}

func ifVersionIsGenesis(asserts ...func([]byte) error) func([]byte) error {
	return ifVersionIs(Genesis, asserts...)
}

func ifVersionIsBasic(asserts ...func([]byte) error) func([]byte) error {
	return ifVersionIs(Basic, asserts...)
}

func ifVersionIs(v uint8, asserts ...func([]byte) error) func([]byte) error {
	return func(b []byte) error {
		if b[0] == v {
			return runAsserts(b, asserts)
		}

		return nil
	}
}

func ifVersionIsCreateOrUpdateStruct(asserts ...func([]byte) error) func([]byte) error {
	return func(b []byte) error {
		switch b[0] {
		case CreateStructure, UpdateStructure:
			return runAsserts(b, asserts)
		}

		return nil
	}
}

func ifVersionIsUpdateAddress(asserts ...func([]byte) error) func([]byte) error {
	return func(b []byte) error {
		switch b[0] {
		case UpdateProfitAddress, UpdateFeeAddress, CreateTransitAddress, DeleteTransitAddress:
			return runAsserts(b, asserts)
		}

		return nil
	}
}

func lengthIs(l int) func([]byte) error {
	return func(b []byte) error {
		if b == nil || len(b) != l {
			return ErrInvalidLength
		}

		return nil
	}
}

func lengthIsValid(b []byte) error {
	currentLen := len(b)
	minimalLen := HeaderLength + TxLength

	if currentLen < minimalLen {
		return ErrInvalidLength
	}

	expectedLen := HeaderLength + (TxLength * int((Block)(b).TxCount()))
	if currentLen != expectedLen {
		return ErrInvalidLength
	}

	return nil
}

func signatureIsValid(b []byte) error {
	pub, msg, sig := b[3:35], b[0:85], b[85:149]

	if len(b) != TxLength {
		pub, msg, sig = b[71:103], b[0:103], b[103:167]
	}

	if !ed25519.Verify(pub, msg, sig) {
		return ErrInvalidSignature
	}

	return nil
}

func senderPrefixIs(v uint16) func([]byte) error {
	return func(b []byte) error {
		if (Transaction)(b).Sender().Version() != v {
			return ErrInvalidSender
		}

		return nil
	}
}

func senderPrefixValidAndNot(v uint16) func([]byte) error {
	return func(b []byte) error {
		ver := (Transaction)(b).Sender().Version()

		if err := adrVersionIsValid(ver); err != nil {
			return ErrInvalidSender
		}

		if ver == v {
			return ErrInvalidSender
		}

		return nil
	}
}

func senderAndRecipientNotEqual(b []byte) error {
	if bytes.Equal((Transaction)(b).Recipient(), (Transaction)(b).Sender()) {
		return ErrInvalidRecipient
	}

	return nil
}

func recipientPrefixIs(v uint16) func([]byte) error {
	return func(b []byte) error {
		if (Transaction)(b).Recipient().Version() != v {
			return ErrInvalidRecipient
		}

		return nil
	}
}

func recipientPrefixValidAndNot(vs ...uint16) func([]byte) error {
	return func(b []byte) error {
		ver := (Transaction)(b).Recipient().Version()

		if err := adrVersionIsValid(ver); err != nil {
			return ErrInvalidRecipient
		}

		if inArray(ver, vs) {
			return ErrInvalidRecipient
		}

		return nil
	}
}

func inArray(n uint16, vs []uint16) bool {
	for _, v := range vs {
		if n == v {
			return true
		}
	}

	return false
}

func structPrefixValidAndNot(vs ...uint16) func([]byte) error {
	return func(b []byte) error {
		ver := binary.BigEndian.Uint16(b[35:37])

		if err := adrVersionIsValid(ver); err != nil {
			return ErrInvalidPrefix
		}

		if inArray(ver, vs) {
			return ErrInvalidPrefix
		}

		return nil
	}
}

func nameIsValid(b []byte) error {
	const maxLength = 35

	if b[41] > maxLength {
		return ErrInvalidName
	}

	if !utf8.ValidString((Transaction)(b).Name()) {
		return ErrInvalidName
	}

	return nil
}

func feePercentBetween(min, max uint16) func([]byte) error {
	return func(b []byte) error {
		p := (Transaction)(b).FeePercent()
		if notBetween(p, min, max) {
			return ErrInvalidFeePercent
		}

		return nil
	}
}

func profitPercentBetween(min, max uint16) func([]byte) error {
	return func(b []byte) error {
		p := (Transaction)(b).ProfitPercent()
		if notBetween(p, min, max) {
			return ErrInvalidProfitPercent
		}

		return nil
	}
}

func notBetween(v, min, max uint16) bool {
	if v < min || v > max {
		return true
	}

	return false
}

func versionIsValid(b []byte) error {
	switch len(b) {
	case AddressLength:
		return adrVersionIsValid(binary.BigEndian.Uint16(b[0:2]))
	case TxLength:
		return txVersionIsValid(b[0])
	default:
		return blkVersionIsValid(b[0])
	}
}

func adrVersionIsValid(v uint16) error {
	if v == genesis {
		return nil
	}

	for i := 0; i < 3; i++ {
		chr := getChrFromVer(v, i)
		if chr < 97 || chr > 122 {
			return ErrInvalidPrefix
		}
	}

	return nil
}

func txVersionIsValid(v uint8) error {
	if v > DeleteTransitAddress {
		return ErrInvalidVersion
	}

	return nil
}

func blkVersionIsValid(v uint8) error {
	if v > Basic {
		return ErrInvalidVersion
	}

	return nil
}

func merkleRootIsValid(b []byte) error {
	mrk, err := calculateMerkleRoot(b)
	if err != nil {
		return err
	}

	if !bytes.Equal((Block)(b).MerkleRootHash(), mrk) {
		return ErrInvalidMerkle
	}

	return nil
}

func prevBlockHashIsNull(b []byte) error {
	if !bytes.Equal((Block)(b).PreviousBlockHash(), make([]byte, 32)) {
		return ErrInvalidPrevHash
	}

	return nil
}

func prevBlockHashNotNull(b []byte) error {
	if bytes.Equal((Block)(b).PreviousBlockHash(), make([]byte, 32)) {
		return ErrInvalidPrevHash
	}

	return nil
}

func allTransactionAreGenesis(b []byte) error {
	for i, l := HeaderLength, len(b); i < l; i += TxLength {
		if b[i] != Genesis {
			return ErrInvalidTx
		}
	}

	return nil
}

func allTransactionNotGenesis(b []byte) error {
	for i, l := HeaderLength, len(b); i < l; i += TxLength {
		if b[i] == Genesis {
			return ErrInvalidTx
		}
	}

	return nil
}

func allTransactionsAreValid(b []byte) (err error) {
	c := fillQueue(b)

	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		go func() {
			validateQueue(c, &err)
			wg.Done()
		}()
	}

	wg.Wait()

	return err
}

func fillQueue(b []byte) <-chan Transaction {
	blk := (Block)(b)
	n := blk.TxCount()
	c := make(chan Transaction, n)

	for i, l := uint16(0), n; i < l; i++ {
		c <- blk.Transaction(i)
	}

	close(c)

	return c
}

func validateQueue(c <-chan Transaction, err *error) {
	for tx := range c {
		if tx.Verify() != nil {
			*err = ErrInvalidTx

			return
		}
	}
}
