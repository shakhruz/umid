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

package ledger

import (
	"gitlab.com/umitop/umid/pkg/umi"
)

type ConfirmerLegacy struct {
	Confirmer
}

func NewConfirmerLegacy(ledger *Ledger) *ConfirmerLegacy {
	return &ConfirmerLegacy{
		Confirmer{
			ledger: ledger,
		},
	}
}

func (confirmer *ConfirmerLegacy) AppendBlockLegacy(blockLegacyRaw []byte) error {
	confirmer.Lock()

	block, err := confirmer.ProcessBlockLegacy(blockLegacyRaw)

	confirmer.Unlock()

	if err == nil {
		err = confirmer.AppendBlock(block)
	}

	return err
}

func (confirmer *ConfirmerLegacy) ProcessBlockLegacy(blockLegacyRaw []byte) (umi.Block, error) {
	confirmer.ResetState()

	if err := confirmer.verifyBlock(blockLegacyRaw); err != nil {
		return nil, err
	}

	blockLegacy := (umi.BlockLegacy)(blockLegacyRaw)

	confirmer.BlockHash = blockLegacy.Hash()
	confirmer.BlockTimestamp = blockLegacy.Timestamp()
	confirmer.BlockHeight++

	block := make(umi.Block, umi.HdrLength)

	// Copy block header.
	copy(block, blockLegacy[:umi.HdrLength])

	handlers := map[string]func(umi.Transaction) (umi.Transaction, error){
		umi.TxGenesis:             confirmer.processGenesisLegacy,
		umi.TxSend:                confirmer.ProcessSendLegacy,
		umi.TxCreateStructure:     confirmer.ProcessCreateStructureLegacy,
		umi.TxUpdateStructure:     confirmer.ProcessUpdateStructureLegacy,
		umi.TxChangeProfitAddress: confirmer.ProcessChangeProfitAddressLegacy,
		umi.TxChangeFeeAddress:    confirmer.ProcessChangeFeeAddressLegacy,
		umi.TxActivateTransit:     confirmer.ProcessActivateTransitLegacy,
		umi.TxDeactivateTransit:   confirmer.ProcessDeactivateTransitLegacy,
	}

	for txIndex, txCount := 0, block.TransactionCount(); txIndex < txCount; txIndex++ {
		confirmer.TransactionHeight++

		transaction := blockLegacy.Transaction(txIndex)

		transaction.SetBlockTimestamp(confirmer.BlockTimestamp)
		transaction.SetBlockHeight(confirmer.BlockHeight)
		transaction.SetBlockTransactionIndex(txIndex)
		transaction.SetTransactionHeight(confirmer.TransactionHeight)

		handler, ok := handlers[transaction.Type()]
		if !ok {
			return nil, errUnsupportedTxType
		}

		transaction, err := handler(transaction)
		if err != nil {
			return nil, err
		}

		block = append(block, transaction...)
	}

	return block, nil
}

func (confirmer *ConfirmerLegacy) processGenesisLegacy(transaction umi.Transaction) (umi.Transaction, error) {
	if err := confirmer.processGenesis(transaction); err != nil {
		return nil, err
	}

	confirmer.setTxRecipient(transaction, transaction.Recipient())

	return transaction, nil
}

func (confirmer *ConfirmerLegacy) ProcessSendLegacy(transaction umi.Transaction) (umi.Transaction, error) {
	if err := confirmer.processSend(transaction); err != nil {
		return nil, err
	}

	sender := transaction.Sender()
	recipient := transaction.Recipient()

	confirmer.setTxSender(transaction, sender)
	confirmer.setTxRecipient(transaction, recipient)

	if feeAmount, feeAddress, ok := confirmer.calculateFee(sender, recipient, transaction.Amount()); ok {
		confirmer.setTxFee(transaction, feeAmount, feeAddress)
	}

	return transaction, nil
}

func (confirmer *ConfirmerLegacy) ProcessCreateStructureLegacy(transaction umi.Transaction) (umi.Transaction, error) {
	if err := confirmer.processCreateStructure(transaction); err != nil {
		return nil, err
	}

	confirmer.setTxSender(transaction, transaction.Sender())

	return transaction, nil
}

func (confirmer *ConfirmerLegacy) ProcessUpdateStructureLegacy(transaction umi.Transaction) (umi.Transaction, error) {
	if err := confirmer.processUpdateStructure(transaction); err != nil {
		return nil, err
	}

	confirmer.setTxSender(transaction, transaction.Sender())

	return transaction, nil
}

func (confirmer *ConfirmerLegacy) ProcessChangeProfitAddressLegacy(
	transaction umi.Transaction) (umi.Transaction, error) {
	if err := confirmer.processChangeProfitAddress(transaction); err != nil {
		return nil, err
	}

	confirmer.setTxSender(transaction, transaction.Sender())
	confirmer.setTxRecipient(transaction, transaction.Recipient())

	return transaction, nil
}

func (confirmer *ConfirmerLegacy) ProcessChangeFeeAddressLegacy(transaction umi.Transaction) (umi.Transaction, error) {
	if err := confirmer.processChangeFeeAddress(transaction); err != nil {
		return nil, err
	}

	confirmer.setTxSender(transaction, transaction.Sender())
	confirmer.setTxRecipient(transaction, transaction.Recipient())

	return transaction, nil
}

func (confirmer *ConfirmerLegacy) ProcessActivateTransitLegacy(transaction umi.Transaction) (umi.Transaction, error) {
	if err := confirmer.processActivateTransit(transaction); err != nil {
		return nil, err
	}

	confirmer.setTxSender(transaction, transaction.Sender())
	confirmer.setTxRecipient(transaction, transaction.Recipient())

	return transaction, nil
}

func (confirmer *ConfirmerLegacy) ProcessDeactivateTransitLegacy(transaction umi.Transaction) (umi.Transaction, error) {
	if err := confirmer.processDeactivateTransit(transaction); err != nil {
		return nil, err
	}

	confirmer.setTxSender(transaction, transaction.Sender())
	confirmer.setTxRecipient(transaction, transaction.Recipient())

	return transaction, nil
}

// setTxRecipient добавляет в подтвержденную транзакцию мета-данные отправителя.
func (confirmer *Confirmer) setTxSender(transaction umi.Transaction, sender umi.Address) {
	senderAccount, _ := confirmer.Account(sender)

	transaction.SetSenderAccountType(senderAccount.Type)
	transaction.SetSenderAccountBalance(senderAccount.Balance)
	transaction.SetSenderAccountInterestRate(senderAccount.InterestRate)
	transaction.SetSenderAccountTransactionCount(senderAccount.TransactionCount)
}

// setTxRecipient добавляет в подтвержденную транзакцию мета-данные получателя.
func (confirmer *Confirmer) setTxRecipient(transaction umi.Transaction, recipient umi.Address) {
	recipientAccount, _ := confirmer.Account(recipient)

	transaction.SetRecipientAccountType(recipientAccount.Type)
	transaction.SetRecipientAccountBalance(recipientAccount.Balance)
	transaction.SetRecipientAccountInterestRate(recipientAccount.InterestRate)
	transaction.SetRecipientAccountTransactionCount(recipientAccount.TransactionCount)
}

// setTxRecipient добавляет в подтвержденную транзакцию мета-данные связанные с комиссией.
func (confirmer *Confirmer) setTxFee(transaction umi.Transaction, feeAmount uint64, feeAddress umi.Address) {
	structure, _ := confirmer.Structure(feeAddress.Prefix())
	feeAccount, _ := confirmer.Account(feeAddress)

	transaction.SetFeeAmount(feeAmount)
	transaction.SetFeePercentMeta(structure.FeePercent)
	transaction.SetFeeAddress(feeAddress)
	transaction.SetFeeAccountBalance(feeAccount.Balance)
	transaction.SetFeeAccountInterestRate(feeAccount.InterestRate)
	transaction.SetFeeAccountTransactionCount(feeAccount.TransactionCount)
}
