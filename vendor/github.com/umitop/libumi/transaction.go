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
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"
	"unicode/utf8"
)

const (
	// TransactionLength ...
	TransactionLength = 150

	umi     = "umi"
	genesis = "genesis"
)

// Transaction types.
const (
	Genesis = iota
	Basic
	CreateStructure
	UpdateStructure
	UpdateProfitAddress
	UpdateFeeAddress
	CreateTransitAddress
	DeleteTransitAddress
)

// Errors.
var (
	ErrInvalidVersion       = errors.New("transaction: invalid version")
	ErrInvalidSender        = errors.New("transaction: invalid sender")
	ErrInvalidValue         = errors.New("transaction: invalid value")
	ErrInvalidRecipient     = errors.New("transaction: invalid recipient")
	ErrInvalidPrefix        = errors.New("transaction: invalid prefix")
	ErrInvalidName          = errors.New("transaction: invalid name")
	ErrInvalidFeePercent    = errors.New("transaction: invalid fee percent")
	ErrInvalidProfitPercent = errors.New("transaction: invalid profit percent")
	ErrInvalidSignature     = errors.New("transaction: invalid signature")
)

// Transaction ...
type Transaction []byte

// NewTransaction ...
func NewTransaction() Transaction {
	t := make(Transaction, TransactionLength)
	_ = t.SetVersion(Basic)

	return t
}

// FeePercent ...
func (t Transaction) FeePercent() uint16 {
	return binary.BigEndian.Uint16(t[39:41])
}

// SetFeePercent ...
func (t Transaction) SetFeePercent(p uint16) (err error) {
	binary.BigEndian.PutUint16(t[39:41], p)

	return err
}

// Hash ...
func (t Transaction) Hash() []byte {
	h := sha256.Sum256(t)

	return h[:]
}

// Name ...
func (t Transaction) Name() string {
	s := make([]byte, t[41])
	copy(s, t[42:(42+t[41])])

	return string(s)
}

// SetName ...
func (t Transaction) SetName(s string) (err error) {
	if len(s) > 35 {
		return ErrInvalidName
	}

	if !utf8.ValidString(s) {
		return ErrInvalidName
	}

	copy(t[42:(42+len(s))], s)
	copy(t[(42+len(s)):76], make([]byte, 35-len(s))) // wipe
	t[41] = uint8(len(s))

	return err
}

// Nonce ...
func (t Transaction) Nonce() uint64 {
	return binary.BigEndian.Uint64(t[77:85])
}

// SetNonce ...
func (t Transaction) SetNonce(v uint64) {
	binary.BigEndian.PutUint64(t[77:85], v)
}

// Prefix ...
func (t Transaction) Prefix() string {
	p := make([]byte, 3)
	p[0] = ((t[35] >> 2) & 31) + 96
	p[1] = (((t[35] & 3) << 3) | (t[36] >> 5)) + 96
	p[2] = (t[36] & 31) + 96

	return string(p)
}

// SetPrefix ...
func (t Transaction) SetPrefix(p string) (err error) {
	if len(p) != 3 {
		return ErrInvalidPrefix
	}

	for _, c := range p {
		if c < 97 || c > 122 {
			return ErrInvalidPrefix
		}
	}

	t[35] = (((p[0] - 96) & 31) << 2) | (((p[1] - 96) & 31) >> 3)
	t[36] = ((p[1] - 96) << 5) | ((p[2] - 96) & 31)

	return err
}

// ProfitPercent ...
func (t Transaction) ProfitPercent() uint16 {
	return binary.BigEndian.Uint16(t[37:39])
}

// SetProfitPercent ...
func (t Transaction) SetProfitPercent(v uint16) (err error) {
	binary.BigEndian.PutUint16(t[37:39], v)

	return err
}

// Recipient ...
func (t Transaction) Recipient() Address {
	return Address(t[35:69])
}

// SetRecipient ...
func (t Transaction) SetRecipient(a Address) {
	copy(t[35:69], a)
}

// Sender ...
func (t Transaction) Sender() Address {
	return Address(t[1:35])
}

// SetSender ...
func (t Transaction) SetSender(a Address) {
	copy(t[1:35], a)
}

// Signature ...
func (t Transaction) Signature() []byte {
	return t[85:149]
}

// SetSignature ...
func (t Transaction) SetSignature(s []byte) (err error) {
	if len(s) != ed25519.SignatureSize {
		return ErrInvalidSignature
	}

	copy(t[85:149], s)
	t[149] = 0 // wipe

	return err
}

// Value ...
func (t Transaction) Value() uint64 {
	return binary.BigEndian.Uint64(t[69:77])
}

// SetValue ...
func (t Transaction) SetValue(v uint64) {
	binary.BigEndian.PutUint64(t[69:77], v)
}

// Version ...
func (t Transaction) Version() uint8 {
	return t[0]
}

// SetVersion ...
func (t Transaction) SetVersion(v uint8) (err error) {
	t[0] = v

	return err
}

// Sign ...
func (t Transaction) Sign(k ed25519.PrivateKey) {
	t.SetNonce(uint64(time.Now().UnixNano()))
	_ = t.SetSignature(ed25519.Sign(k, t[0:85]))
}

// Verify ...
func (t Transaction) Verify() (err error) {
	err = t.verifyVersion(err)
	err = t.verifySender(err)

	if t.Version() == CreateStructure || t.Version() == UpdateStructure {
		err = t.verifyPrefix(err)
		err = t.verifyName(err)
		err = t.verifyProfitPercent(err)
		err = t.verifyFeePercent(err)
	} else {
		err = t.verifyRecipient(err)
	}

	if t.Version() == Genesis || t.Version() == Basic {
		err = t.verifyValue(err)
	}

	err = t.verifySignature(err)

	return err
}

func (t Transaction) verifyVersion(err error) error {
	if t.Version() > DeleteTransitAddress {
		err = ErrInvalidVersion
	}

	return err
}

func (t Transaction) verifySender(err error) error {
	if err != nil {
		return err
	}

	switch t.Version() {
	case Genesis:
		if t.Sender().Prefix() != genesis {
			err = ErrInvalidSender
		}
	case Basic:
		if t.Sender().Prefix() == genesis {
			err = ErrInvalidSender
		}
	default:
		if t.Sender().Prefix() != umi {
			err = ErrInvalidSender
		}
	}

	return err
}

func (t Transaction) verifyRecipient(err error) error {
	if err != nil {
		return err
	}

	// нельзя отправлять самому себе
	if bytes.Equal(t.Sender(), t.Recipient()) {
		return ErrInvalidRecipient
	}

	// нельзя отправлять на генезис-адрес
	if t.Recipient().Prefix() == genesis {
		return ErrInvalidRecipient
	}

	if t.Version() != Basic && t.Recipient().Prefix() == umi {
		return ErrInvalidRecipient
	}

	return err
}

func (t Transaction) verifyValue(err error) error {
	if err != nil {
		return err
	}

	if t.Value() > 90_071_992_547_409_91 {
		return ErrInvalidValue
	}

	return err
}

func (t Transaction) verifyPrefix(err error) error {
	if err != nil {
		return err
	}

	pfx := t.Prefix()

	for _, c := range pfx {
		if c < 97 || c > 122 {
			return ErrInvalidPrefix
		}
	}

	if pfx == umi {
		err = ErrInvalidPrefix
	}

	return err
}

func (t Transaction) verifyName(err error) error {
	if err != nil {
		return err
	}

	if t[41] > 35 {
		err = ErrInvalidName
	}

	if !utf8.Valid(t[42 : 42+t[41]]) {
		err = ErrInvalidName
	}

	return err
}

func (t Transaction) verifyProfitPercent(err error) error {
	if err != nil {
		return err
	}

	prf := t.ProfitPercent()
	if prf < 100 || prf > 500 {
		err = ErrInvalidProfitPercent
	}

	return err
}

func (t Transaction) verifyFeePercent(err error) error {
	if err != nil {
		return err
	}

	if t.FeePercent() > 2000 {
		err = ErrInvalidFeePercent
	}

	return err
}

func (t Transaction) verifySignature(err error) error {
	if err != nil {
		return err
	}

	if !ed25519.Verify(t.Sender().PublicKey(), t[0:85], t.Signature()) {
		err = ErrInvalidSignature
	}

	return err
}
