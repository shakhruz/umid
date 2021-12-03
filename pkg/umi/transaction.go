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
	"encoding/binary"
	"encoding/json"
	"time"
)

const (
	TxLength          = 150
	TxConfirmedLength = 268
)

const (
	TxV0Genesis uint8 = iota
	TxV1Send
	TxV2CreateStructure
	TxV3UpdateStructure
	TxV4ChangeProfitAddress
	TxV5ChangeFeeAddress
	TxV6ActivateTransit
	TxV7DeactivateTransit
	TxV8Send
	TxV9CreateStructure
	TxV10UpdateStructure
	TxV11ChangeProfitAddress
	TxV12ChangeFeeAddress
	TxV13ActivateTransit
	TxV14DeactivateTransit
	TxV15Burn
)

const (
	TxGenesis             = "genesis"
	TxSend                = "send"
	TxCreateStructure     = "createStructure"
	TxUpdateStructure     = "updateStructure"
	TxChangeProfitAddress = "changeProfitAddress"
	TxChangeFeeAddress    = "changeFeeAddress"
	TxActivateTransit     = "activateTransit"
	TxDeactivateTransit   = "deactivateTransit"
	TxBurn                = "burn"
	txUnknown             = "unknown"
)

type Transaction []byte

func NewTransaction() Transaction {
	return make(Transaction, TxLength)
}

func (transaction Transaction) Hash() Hash {
	return sha256.Sum256(transaction[0:150])
}

func (transaction Transaction) Type() string { //nolint:revive,cyclop // Easy to read and understand
	switch transaction[0] {
	case TxV0Genesis:
		return TxGenesis
	case TxV1Send, TxV8Send:
		return TxSend
	case TxV2CreateStructure, TxV9CreateStructure:
		return TxCreateStructure
	case TxV3UpdateStructure, TxV10UpdateStructure:
		return TxUpdateStructure
	case TxV4ChangeProfitAddress, TxV11ChangeProfitAddress:
		return TxChangeProfitAddress
	case TxV5ChangeFeeAddress, TxV12ChangeFeeAddress:
		return TxChangeFeeAddress
	case TxV6ActivateTransit, TxV13ActivateTransit:
		return TxActivateTransit
	case TxV7DeactivateTransit, TxV14DeactivateTransit:
		return TxDeactivateTransit
	case TxV15Burn:
		return TxBurn
	default:
		return txUnknown
	}
}

func (transaction Transaction) Version() uint8 {
	return transaction[0]
}

func (transaction Transaction) SetVersion(version uint8) Transaction {
	transaction[0] = version

	return transaction
}

func (transaction Transaction) Sender() (sender Address) {
	copy(sender[0:34], transaction[1:35])

	return sender
}

func (transaction Transaction) SetSender(sender Address) Transaction {
	copy(transaction[1:35], sender[0:34])

	return transaction
}

func (transaction Transaction) Recipient() (recipient Address) {
	copy(recipient[0:34], transaction[35:69])

	return recipient
}

func (transaction Transaction) SetRecipient(recipient Address) Transaction {
	copy(transaction[35:69], recipient[0:34])

	return transaction
}

func (transaction Transaction) Amount() uint64 {
	switch transaction.Type() {
	case TxGenesis, TxSend, TxBurn:
		return binary.BigEndian.Uint64(transaction[69:77])

	case TxCreateStructure:
		return 50_000_00

	default:
		return 0
	}
}

func (transaction Transaction) SetAmount(amount uint64) Transaction {
	binary.BigEndian.PutUint64(transaction[69:77], amount)

	return transaction
}

func (transaction Transaction) Timestamp() uint32 {
	return binary.BigEndian.Uint32(transaction[77:81])
}

func (transaction Transaction) SetTimestamp(epoch uint32) Transaction {
	binary.BigEndian.PutUint32(transaction[77:81], epoch)

	return transaction
}

func (transaction Transaction) Nonce() uint32 {
	return binary.BigEndian.Uint32(transaction[81:85])
}

func (transaction Transaction) SetNonce(nonce uint32) Transaction {
	binary.BigEndian.PutUint32(transaction[81:85], nonce)

	return transaction
}

func (transaction Transaction) Prefix() Prefix {
	return (Prefix)(binary.BigEndian.Uint16(transaction[35:37]))
}

func (transaction Transaction) SetPrefix(pfx Prefix) Transaction {
	binary.BigEndian.PutUint16(transaction[35:37], (uint16)(pfx))

	return transaction
}

func (transaction Transaction) ProfitPercent() uint16 {
	return binary.BigEndian.Uint16(transaction[37:39])
}

func (transaction Transaction) SetProfitPercent(val uint16) Transaction {
	binary.BigEndian.PutUint16(transaction[37:39], val)

	return transaction
}

func (transaction Transaction) FeePercent() uint16 {
	return binary.BigEndian.Uint16(transaction[39:41])
}

func (transaction Transaction) SetFeePercent(val uint16) Transaction {
	binary.BigEndian.PutUint16(transaction[39:41], val)

	return transaction
}

func (transaction Transaction) Description() (description string) {
	strLen := int(transaction[41])
	strLow := 42
	strHigh := strLow + strLen

	description = string(transaction[strLow:strHigh])

	return description
}

func (transaction Transaction) SetDescription(description string) Transaction {
	strLen := len(description)
	strLow := 42
	strHigh := strLow + strLen

	transaction[41] = uint8(strLen)
	copy(transaction[strLow:strHigh], description)

	return transaction
}

// meta

func (transaction Transaction) BlockTimestamp() uint32 {
	return binary.BigEndian.Uint32(transaction[150:154])
}

func (transaction Transaction) SetBlockTimestamp(timestamp uint32) {
	binary.BigEndian.PutUint32(transaction[150:154], timestamp)
}

func (transaction Transaction) BlockHeight() uint32 {
	return binary.BigEndian.Uint32(transaction[154:158])
}

func (transaction Transaction) SetBlockHeight(height uint32) {
	binary.BigEndian.PutUint32(transaction[154:158], height)
}

func (transaction Transaction) BlockTransactionIndex() uint16 {
	return binary.BigEndian.Uint16(transaction[158:160])
}

func (transaction Transaction) SetBlockTransactionIndex(index int) {
	binary.BigEndian.PutUint16(transaction[158:160], uint16(index))
}

func (transaction Transaction) TransactionHeight() uint64 {
	return binary.BigEndian.Uint64(transaction[160:168])
}

func (transaction Transaction) SetTransactionHeight(height uint64) {
	binary.BigEndian.PutUint64(transaction[160:168], height)
}

// meta - sender

func (transaction Transaction) SenderAccountType() AccountType {
	return (AccountType)(transaction[168])
}

func (transaction Transaction) SetSenderAccountType(accountType AccountType) {
	transaction[168] = uint8(accountType)
}

func (transaction Transaction) SenderAccountBalance() uint64 {
	return binary.BigEndian.Uint64(transaction[169:177])
}

func (transaction Transaction) SetSenderAccountBalance(balance uint64) {
	binary.BigEndian.PutUint64(transaction[169:177], balance)
}

func (transaction Transaction) SenderAccountInterestRate() uint16 {
	return binary.BigEndian.Uint16(transaction[177:179])
}

func (transaction Transaction) SetSenderAccountInterestRate(percent uint16) {
	binary.BigEndian.PutUint16(transaction[177:179], percent)
}

func (transaction Transaction) SenderAccountTransactionCount() uint64 {
	return binary.BigEndian.Uint64(transaction[179:187])
}

func (transaction Transaction) SetSenderAccountTransactionCount(height uint64) {
	binary.BigEndian.PutUint64(transaction[179:187], height)
}

// meta - recipient

func (transaction Transaction) RecipientAccountType() AccountType {
	return (AccountType)(transaction[187])
}

func (transaction Transaction) SetRecipientAccountType(accountType AccountType) {
	transaction[187] = uint8(accountType)
}

func (transaction Transaction) RecipientAccountBalance() uint64 {
	return binary.BigEndian.Uint64(transaction[188:196])
}

func (transaction Transaction) SetRecipientAccountBalance(balance uint64) {
	binary.BigEndian.PutUint64(transaction[188:196], balance)
}

func (transaction Transaction) RecipientAccountInterestRate() uint16 {
	return binary.BigEndian.Uint16(transaction[196:198])
}

func (transaction Transaction) SetRecipientAccountInterestRate(percent uint16) {
	binary.BigEndian.PutUint16(transaction[196:198], percent)
}

func (transaction Transaction) RecipientAccountTransactionCount() uint64 {
	return binary.BigEndian.Uint64(transaction[198:206])
}

func (transaction Transaction) SetRecipientAccountTransactionCount(height uint64) {
	binary.BigEndian.PutUint64(transaction[198:206], height)
}

// meta - fee

func (transaction Transaction) FeeAmount() uint64 {
	return binary.BigEndian.Uint64(transaction[206:214])
}

func (transaction Transaction) SetFeeAmount(amount uint64) {
	binary.BigEndian.PutUint64(transaction[206:214], amount)
}

func (transaction Transaction) FeePercentMeta() uint16 {
	return binary.BigEndian.Uint16(transaction[214:216])
}

func (transaction Transaction) SetFeePercentMeta(percent uint16) {
	binary.BigEndian.PutUint16(transaction[214:216], percent)
}

func (transaction Transaction) FeeAddress() (address Address) {
	copy(address[0:34], transaction[216:250])

	return address
}

func (transaction Transaction) SetFeeAddress(address Address) {
	copy(transaction[216:250], address[0:34])
}

func (transaction Transaction) FeeAccountBalance() uint64 {
	return binary.BigEndian.Uint64(transaction[250:258])
}

func (transaction Transaction) SetFeeAccountBalance(balance uint64) {
	binary.BigEndian.PutUint64(transaction[250:258], balance)
}

func (transaction Transaction) FeeAccountInterestRate() uint16 {
	return binary.BigEndian.Uint16(transaction[258:260])
}

func (transaction Transaction) SetFeeAccountInterestRate(percent uint16) {
	binary.BigEndian.PutUint16(transaction[258:260], percent)
}

func (transaction Transaction) FeeAccountTransactionCount() uint64 {
	return binary.BigEndian.Uint64(transaction[260:268])
}

func (transaction Transaction) SetFeeAccountTransactionCount(height uint64) {
	binary.BigEndian.PutUint64(transaction[260:268], height)
}

// extra

func (transaction Transaction) HasRecipient() bool {
	switch transaction.Type() {
	case TxCreateStructure, TxUpdateStructure, TxBurn:
		return false
	default:
		return true
	}
}

func (transaction Transaction) HasFee() bool {
	return len(transaction) == TxConfirmedLength && transaction[216] != 0
}

// JSON

//nolint:funlen,revive // ...
func (transaction Transaction) MarshalJSON() ([]byte, error) {
	data := struct {
		Height                *uint64 `json:"height,omitempty"`
		BlockTimestamp        *string `json:"blockTimestamp,omitempty"`
		BlockHeight           *uint32 `json:"blockHeight,omitempty"`
		BlockTransactionIndex *uint16 `json:"blockTransactionIndex,omitempty"`

		Hash    string `json:"hash"`
		Type    string `json:"type"`
		Version uint8  `json:"version"`

		Amount uint64 `json:"amount"`

		SenderAddress                 string  `json:"senderAddress"`
		SenderAccountType             *string `json:"senderAccountType,omitempty"`
		SenderAccountBalance          *uint64 `json:"senderAccountBalance,omitempty"`
		SenderAccountInterestRate     *uint16 `json:"senderAccountInterestRate,omitempty"`
		SenderAccountTransactionCount *uint64 `json:"senderAccountTransactionCount,omitempty"`

		RecipientAmount                  *uint64 `json:"recipientAmount,omitempty"`
		RecipientAddress                 *string `json:"recipientAddress,omitempty"`
		RecipientAccountType             *string `json:"recipientAccountType,omitempty"`
		RecipientAccountBalance          *uint64 `json:"recipientAccountBalance,omitempty"`
		RecipientAccountInterestRate     *uint16 `json:"recipientAccountInterestRate,omitempty"`
		RecipientAccountTransactionCount *uint64 `json:"recipientAccountTransactionCount,omitempty"`

		FeeAmount                  *uint64 `json:"feeAmount,omitempty"`
		FeeAddress                 *string `json:"feeAddress,omitempty"`
		FeeAccountBalance          *uint64 `json:"feeAccountBalance,omitempty"`
		FeeAccountInterestRate     *uint16 `json:"feeAccountInterestRate,omitempty"`
		FeeAccountTransactionCount *uint64 `json:"feeAccountTransactionCount,omitempty"`

		Prefix        *string `json:"prefix,omitempty"`
		Description   *string `json:"description,omitempty"`
		ProfitPercent *uint16 `json:"profitPercent,omitempty"`
		FeePercent    *uint16 `json:"feePercent,omitempty"`

		Timestamp *string `json:"timestamp,omitempty"`
	}{
		Hash:          transaction.Hash().String(),
		Type:          transaction.Type(),
		Version:       transaction.Version(),
		SenderAddress: transaction.Sender().String(),
		Amount:        transaction.Amount(),
	}

	if len(transaction) == TxConfirmedLength {
		data.Height = new(uint64)
		*data.Height = transaction.TransactionHeight()

		data.BlockTimestamp = new(string)
		*data.BlockTimestamp = time.Unix(int64(transaction.BlockTimestamp()), 0).UTC().Format(time.RFC3339)

		data.BlockHeight = new(uint32)
		*data.BlockHeight = transaction.BlockHeight()

		data.BlockTransactionIndex = new(uint16)
		*data.BlockTransactionIndex = transaction.BlockTransactionIndex()

		data.SenderAccountType = new(string)
		*data.SenderAccountType = transaction.SenderAccountType().String()

		data.SenderAccountBalance = new(uint64)
		*data.SenderAccountBalance = transaction.SenderAccountBalance()

		data.SenderAccountInterestRate = new(uint16)
		*data.SenderAccountInterestRate = transaction.SenderAccountInterestRate()

		data.SenderAccountTransactionCount = new(uint64)
		*data.SenderAccountTransactionCount = transaction.SenderAccountTransactionCount()
	}

	if transaction.HasRecipient() {
		data.RecipientAddress = new(string)
		*data.RecipientAddress = transaction.Recipient().String()

		if len(transaction) == TxConfirmedLength {
			data.RecipientAmount = new(uint64)
			*data.RecipientAmount = transaction.Amount() - transaction.FeeAmount()

			data.RecipientAccountType = new(string)
			*data.RecipientAccountType = transaction.RecipientAccountType().String()

			data.RecipientAccountBalance = new(uint64)
			*data.RecipientAccountBalance = transaction.RecipientAccountBalance()

			data.RecipientAccountInterestRate = new(uint16)
			*data.RecipientAccountInterestRate = transaction.RecipientAccountInterestRate()

			data.RecipientAccountTransactionCount = new(uint64)
			*data.RecipientAccountTransactionCount = transaction.RecipientAccountTransactionCount()
		}
	}

	if len(transaction) == TxConfirmedLength && transaction.HasFee() {
		data.FeeAmount = new(uint64)
		*data.FeeAmount = transaction.FeeAmount()

		data.FeePercent = new(uint16)
		*data.FeePercent = transaction.FeePercentMeta()

		data.FeeAddress = new(string)
		*data.FeeAddress = transaction.FeeAddress().String()

		data.FeeAccountBalance = new(uint64)
		*data.FeeAccountBalance = transaction.FeeAccountBalance()

		data.FeeAccountInterestRate = new(uint16)
		*data.FeeAccountInterestRate = transaction.FeeAccountInterestRate()

		data.FeeAccountTransactionCount = new(uint64)
		*data.FeeAccountTransactionCount = transaction.FeeAccountTransactionCount()
	}

	switch transaction.Type() {
	case TxCreateStructure, TxUpdateStructure:
		data.Prefix = new(string)
		*data.Prefix = transaction.Prefix().String()

		data.Description = new(string)
		*data.Description = transaction.Description()

		data.ProfitPercent = new(uint16)
		*data.ProfitPercent = transaction.ProfitPercent()

		data.FeePercent = new(uint16)
		*data.FeePercent = transaction.FeePercent()
	}

	if transaction.Version() >= TxV8Send {
		data.Timestamp = new(string)
		*data.Timestamp = time.Unix(int64(transaction.Timestamp()), 0).UTC().Format(time.RFC3339)
	}

	return json.Marshal(data) //nolint:wrapcheck // ...
}
