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
	"time"
	"unicode/utf8"
)

// TxLength ...
const TxLength = 150

// Transaction types.
const (
	Genesis uint8 = iota
	Basic
	CreateStructure
	UpdateStructure
	UpdateProfitAddress
	UpdateFeeAddress
	CreateTransitAddress
	DeleteTransitAddress
)

const (
	verGenesis = 0
	verUmi     = 21929
)

// Errors.
var (
	ErrTxInvalidLength        = errors.New("invalid length")
	ErrTxInvalidVersion       = errors.New("invalid version")
	ErrTxInvalidSender        = errors.New("invalid sender")
	ErrTxInvalidRecipient     = errors.New("invalid recipient")
	ErrTxInvalidPrefix        = errors.New("invalid prefix")
	ErrTxInvalidName          = errors.New("invalid name")
	ErrTxInvalidFeePercent    = errors.New("invalid fee percent")
	ErrTxInvalidProfitPercent = errors.New("invalid profit percent")
	ErrTxInvalidSignature     = errors.New("invalid signature")
)

// VersionTx ...
func VersionTx(t []byte) uint8 {
	return t[0]
}

// SignTx ...
func SignTx(t []byte, sec []byte) {
	setTxNonce(t, uint64(time.Now().UnixNano()))
	setTxSignature(t, ed25519.Sign(sec, t[0:85]))
}

func setTxVersion(t []byte, n uint8) {
	t[0] = n
}

func setTxNonce(t []byte, n uint64) {
	binary.BigEndian.PutUint64(t[77:85], n)
}

func setTxSignature(t []byte, b []byte) {
	copy(t[85:149], b)
}

// VerifyTx ...
func VerifyTx(t []byte) error {
	if len(t) != 150 {
		return ErrTxInvalidLength
	}

	return verifyVersionTx(t)
}

func verifyVersionTx(t []byte) error {
	switch VersionTx(t) {
	case Basic, Genesis:
		return verifyTxBasicAndGenesis(t)
	case CreateStructure, UpdateStructure:
		return verifyTxStruct(t)
	case UpdateProfitAddress, UpdateFeeAddress, CreateTransitAddress, DeleteTransitAddress:
		return verifyTxAddress(t)
	default:
		return ErrTxInvalidVersion
	}
}

func verifyTxBasicAndGenesis(t TxBasic) error {
	if t.Version() == Genesis {
		return verifyTxGenesis(t)
	}

	return verifyTxBasic(t)
}

func verifyTxGenesis(t TxBasic) error {
	if t.Sender().Version() != verGenesis {
		return ErrTxInvalidSender
	}

	if t.Recipient().Version() != verUmi {
		return ErrTxInvalidRecipient
	}

	return verifySignatureTx(t)
}

func verifyTxBasic(t TxBasic) error {
	if bytes.Equal(t.Sender(), t.Recipient()) {
		return ErrTxInvalidRecipient
	}

	if t.Sender().Version() == verGenesis {
		return ErrTxInvalidSender
	}

	if t.Recipient().Version() == verGenesis {
		return ErrTxInvalidRecipient
	}

	return verifySignatureTx(t)
}

func verifyTxStruct(t TxStruct) error {
	if t.Sender().Version() != verUmi {
		return ErrTxInvalidSender
	}

	switch t.Prefix() {
	case "umi", "genesis":
		return ErrTxInvalidPrefix
	}

	return verifyTxStructProfitPercent(t)
}

func verifyTxStructProfitPercent(t TxStruct) error {
	const (
		minProfitPercent = 100
		maxProfitPercent = 500
	)

	if t.ProfitPercent() < minProfitPercent || t.ProfitPercent() > maxProfitPercent {
		return ErrTxInvalidProfitPercent
	}

	return verifyTxStructFeePercent(t)
}

func verifyTxStructFeePercent(t TxStruct) error {
	const maxFeePercent = 2000

	if t.FeePercent() > maxFeePercent {
		return ErrTxInvalidFeePercent
	}

	return verifyTxStructName(t)
}

func verifyTxStructName(t TxStruct) error {
	const maxNameLength = 35

	if t[41] > maxNameLength {
		return ErrTxInvalidName
	}

	if !utf8.ValidString(t.Name()) {
		return ErrTxInvalidName
	}

	return verifySignatureTx(t)
}

func verifyTxAddress(t TxAddress) error {
	if t.Sender().Version() != verUmi {
		return ErrTxInvalidSender
	}

	switch t.Address().Version() {
	case verUmi, verGenesis:
		return ErrTxInvalidRecipient
	}

	return verifySignatureTx(t)
}

func verifySignatureTx(t []byte) error {
	if !ed25519.Verify(t[3:35], t[0:85], t[85:149]) {
		return ErrTxInvalidSignature
	}

	return nil
}
