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
	"crypto/ed25519"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrVerify = errors.New("err")

func (transaction Transaction) Verify() error {
	switch transaction.Type() {
	case TxGenesis:
		return verifyGenesis(transaction)

	case TxSend:
		return verifySend(transaction)

	case TxCreateStructure, TxUpdateStructure:
		return verifyStructure(transaction)

	case TxChangeProfitAddress, TxChangeFeeAddress, TxActivateTransit, TxDeactivateTransit:
		return verifyAddress(transaction)

	case TxBurn:
		return verifyBurn(transaction)

	case TxIssue:
		return verifyIssue(transaction)

	case TxMintNft:
		return verifyMintNft(transaction)

	case TxMintNftWitness:
		return verifyMintNftWitness(transaction)
	}

	return nil
}

func verifySignature(transaction Transaction) bool {
	if transaction[0] < 8 {
		if transaction[149] != 0 {
			return false
		}

		return ed25519.Verify((ed25519.PublicKey)(transaction[3:35]), transaction[0:85], transaction[85:149])
	}

	return ed25519.Verify((ed25519.PublicKey)(transaction[3:35]), transaction[0:86], transaction[86:150])
}

func verifyGenesis(transaction Transaction) error {
	if transaction.Sender().Prefix() != PfxVerGenesis {
		return fmt.Errorf("%w: sender must be 'genesis'", ErrVerify)
	}

	if transaction.Recipient().Prefix() != PfxVerUmi {
		return fmt.Errorf("%w: recipient must be 'umi'", ErrVerify)
	}

	if transaction.Amount() == 0 {
		return fmt.Errorf("%w: amount must not be 0", ErrVerify)
	}

	if !verifySignature(transaction) {
		return fmt.Errorf("%w: invalid signature", ErrVerify)
	}

	return nil
}

func verifySend(transaction Transaction) error {
	sender := transaction.Sender()
	recipient := transaction.Recipient()

	if sender == recipient {
		return fmt.Errorf("%w: sender and recipient must not be equal", ErrVerify)
	}

	if sender.Prefix() == PfxVerGenesis {
		return fmt.Errorf("%w: sender must not be 'genesis'", ErrVerify)
	}

	if recipient.Prefix() == PfxVerGenesis {
		return fmt.Errorf("%w: recipient must not be 'genesis'", ErrVerify)
	}

	if transaction.Version() > 7 && transaction.Amount() == 0 {
		return fmt.Errorf("%w: amount must not be 0", ErrVerify)
	}

	if !verifySignature(transaction) {
		return fmt.Errorf("%w: invalid signature", ErrVerify)
	}

	return nil
}

func verifyStructure(transaction Transaction) error {
	if transaction.Sender().Prefix() != PfxVerUmi {
		return fmt.Errorf("%w: sender must be 'umi'", ErrVerify)
	}

	if !transaction.Prefix().IsValid() {
		return fmt.Errorf("%w: invalid prefix", ErrVerify)
	}

	profitPercent := transaction.ProfitPercent()

	if profitPercent < 1_00 || profitPercent > 5_00 {
		return fmt.Errorf("%w: profit percent must be between 100 and 500", ErrVerify)
	}

	if transaction.FeePercent() > 20_00 {
		return fmt.Errorf("%w: fee percent value must be between 0 and 2000", ErrVerify)
	}

	if transaction[41] > 35 {
		return fmt.Errorf("%w: invalid description length", ErrVerify)
	}

	if !verifySignature(transaction) {
		return fmt.Errorf("%w: invalid signature", ErrVerify)
	}

	return nil
}

func verifyAddress(transaction Transaction) error {
	if transaction.Sender().Prefix() != PfxVerUmi {
		return fmt.Errorf("%w: sender must be 'umi'", ErrVerify)
	}

	switch transaction.Recipient().Prefix() {
	case PfxVerGenesis, PfxVerUmi:
		return fmt.Errorf("%w: recipient must not be 'genesis' and 'umi'", ErrVerify)
	}

	if !verifySignature(transaction) {
		return fmt.Errorf("%w: invalid signature", ErrVerify)
	}

	return nil
}

func verifyBurn(transaction Transaction) error {
	sender := transaction.Sender()

	if sender.Prefix() == PfxVerGenesis {
		return fmt.Errorf("%w: sender must not be 'genesis'", ErrVerify)
	}

	if !verifySignature(transaction) {
		return fmt.Errorf("%w: invalid signature", ErrVerify)
	}

	return nil
}

func verifyIssue(transaction Transaction) error {
	sender := transaction.Sender()
	recipient := transaction.Recipient()

	if sender.Prefix() != PfxVerUmi {
		return fmt.Errorf("%w: sender must be 'umi'", ErrVerify)
	}

	if recipient.Prefix() != PfxVerNft {
		return fmt.Errorf("%w: recipient must be 'nft'", ErrVerify)
	}

	if !verifySignature(transaction) {
		return fmt.Errorf("%w: invalid signature", ErrVerify)
	}

	return nil
}

func verifyMintNft(transaction Transaction) error {
	tx := transaction

	hdrLength := 17  // type (1 byte), timestamp (4 byte), nonce (4 bytes), meta len (4 bytes), data len (4 bytes)
	minLength := 115 // header (17 bytes) + sender (34) + signature (64 bytes)
	txLength := len(tx)

	if txLength < minLength {
		return fmt.Errorf("%w: too short", ErrVerify)
	}

	metaLength := int(binary.BigEndian.Uint32(tx[9:13]))
	dataLength := int(binary.BigEndian.Uint32(tx[13:17]))
	totalLength := minLength + metaLength + dataLength

	if txLength != totalLength {
		return fmt.Errorf("%w: invalid length", ErrVerify)
	}

	var sender Address

	senderOffset := hdrLength + metaLength + dataLength // addr prefix
	sigOffset := senderOffset + AddrLength

	copy(sender[:], tx[senderOffset:sigOffset])

	pub := sender.PublicKey()

	if !ed25519.Verify((ed25519.PublicKey)(pub), tx[:sigOffset], tx[sigOffset:]) {
		return fmt.Errorf("%w: inavlid signature", ErrVerify)
	}

	if metaLength > 0 {
		type MetaSchema struct {
			ContentType *string `json:"contentType"`
		}

		meta := &MetaSchema{}

		if err := json.Unmarshal(tx[hdrLength:hdrLength+metaLength], meta); err != nil {
			return fmt.Errorf("%w: malformed meta", ErrVerify)
		}
	}

	return nil
}

func verifyMintNftWitness(transaction Transaction) error {
	sender := transaction.Sender()

	if sender.Prefix() != PfxVerNft {
		return fmt.Errorf("%w: sender must be 'nft'", ErrVerify)
	}

	return nil
}
