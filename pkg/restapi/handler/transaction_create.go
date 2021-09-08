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

package handler

import (
	"crypto/ed25519"
	"encoding/json"
	"net/http"

	"gitlab.com/umitop/umid/pkg/umi"
)

type CreateTransactionRequest struct {
	Type             *string `json:"type,omitempty"`
	SenderAddress    *string `json:"senderAddress,omitempty"`
	RecipientAddress *string `json:"recipientAddress,omitempty"`
	Amount           *uint64 `json:"amount,omitempty"`
	Prefix           *string `json:"prefix,omitempty"`
	Description      *string `json:"description,omitempty"`
	ProfitPercent    *uint16 `json:"profitPercent,omitempty"`
	FeePercent       *uint16 `json:"feePercent,omitempty"`
	Seed             *[]byte `json:"seed,omitempty"`
}

type CreateTransactionResponse struct {
	Data  []byte `json:"data,omitempty"`
	Error *Error `json:"error,omitempty"`
}

func CreateTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(CreateTransactionResponse)
		response.Data, response.Error = processCreateTransaction(r)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func processCreateTransaction(r *http.Request) ([]byte, *Error) {
	request := new(CreateTransactionRequest)

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, NewError(400, err.Error())
	}

	if err := verifyCreateTransactionRequest(request); err != nil {
		return nil, err
	}

	transaction := buildTransaction(request)

	if err := transaction.Verify(); err != nil {
		return nil, NewError(400, err.Error())
	}

	return transaction, nil
}

func verifyCreateTransactionRequest(request *CreateTransactionRequest) *Error {
	if request.Seed == nil {
		return NewError(-1, "Параметр 'seed' является обязательным.")
	}

	if len(*request.Seed) != 32 {
		return NewError(-1, "Длина 'seed' должна быть 32 байта.")
	}

	if request.Type == nil {
		return NewError(-1, "Параметр 'type' является обязательным.")
	}

	switch *request.Type {
	case umi.TxSend:
		return verifyTxSend(request)

	case umi.TxCreateStructure, umi.TxUpdateStructure:
		return verifyTxStructure(request)

	case umi.TxChangeProfitAddress, umi.TxChangeFeeAddress, umi.TxActivateTransit, umi.TxDeactivateTransit:
		return verifyTxAddress(request)

	default:
		return NewError(-1, "Некорректное значение параметра 'type'.")
	}
}

func verifySender(request *CreateTransactionRequest) *Error {
	if request.SenderAddress == nil {
		return NewError(-1, "Параметр 'senderAddress' является обязательным.")
	}

	if !umi.IsBech32Valid(*request.SenderAddress) {
		return NewError(-1, "Значение параметра 'senderAddress' должно быть валидным адресом в формате bech32.")
	}

	return nil
}

func verifyRecipient(request *CreateTransactionRequest) *Error {
	if request.RecipientAddress == nil {
		return NewError(-1, "Для транзакции имеющий тип 'send' параметр 'recipientAddress' является обязательным.")
	}

	if !umi.IsBech32Valid(*request.RecipientAddress) {
		return NewError(-1, "Значение параметра 'recipientAddress' должно быть валидным адресом в формате bech32.")
	}

	return nil
}

func verifyTxSend(request *CreateTransactionRequest) *Error {
	if err := verifySender(request); err != nil {
		return err
	}

	if err := verifyRecipient(request); err != nil {
		return err
	}

	if request.Amount == nil {
		return NewError(-1, "Для транзакции имеющий тип 'send' параметр 'amount' является обязательным.")
	}

	if *request.Amount == 0 {
		return NewError(-1, "Значение параметра 'amount' должно быть больше нуля.")
	}

	return nil
}

func verifyTxStructure(request *CreateTransactionRequest) *Error {
	if err := verifySender(request); err != nil {
		return err
	}

	if request.Prefix == nil {
		return NewError(-1, "Параметр 'prefix' является обязательным.")
	}

	if !umi.VerifyHrp(*request.Prefix) {
		return NewError(-1, "Параметр 'prefix' может содержать только 3 символа латиницы в нижнем регистре.")
	}

	if request.Description == nil {
		return NewError(-1, "Параметр 'description' является обязательным.")
	}

	if len(*request.Description) > 53 {
		return NewError(-1, "Длина 'description' не может превышать 53 байта.")
	}

	if request.ProfitPercent == nil {
		return NewError(-1, "Параметр 'profitPercent' является обязательным.")
	}

	if request.FeePercent == nil {
		return NewError(-1, "Параметр 'feePercent' является обязательным.")
	}

	return nil
}

func verifyTxAddress(request *CreateTransactionRequest) *Error {
	if err := verifySender(request); err != nil {
		return err
	}

	return verifyRecipient(request)
}

func buildTransaction(request *CreateTransactionRequest) umi.Transaction {
	transaction := umi.NewTransaction()

	sender, _ := umi.ParseAddress(*request.SenderAddress)

	transaction.SetSender(sender)

	switch *request.Type {
	case umi.TxSend:
		recipient, _ := umi.ParseAddress(*request.RecipientAddress)

		transaction.SetVersion(umi.TxV1Send)
		transaction.SetRecipient(recipient)
		transaction.SetAmount(*request.Amount)

	case umi.TxCreateStructure:
		transaction.SetVersion(umi.TxV2CreateStructure)

	case umi.TxUpdateStructure:
		transaction.SetVersion(umi.TxV3UpdateStructure)

	case umi.TxChangeProfitAddress:
		recipient, _ := umi.ParseAddress(*request.RecipientAddress)

		transaction.SetVersion(umi.TxV4ChangeProfitAddress)
		transaction.SetRecipient(recipient)

	case umi.TxChangeFeeAddress:
		recipient, _ := umi.ParseAddress(*request.RecipientAddress)

		transaction.SetVersion(umi.TxV5ChangeFeeAddress)
		transaction.SetRecipient(recipient)

	case umi.TxActivateTransit:
		recipient, _ := umi.ParseAddress(*request.RecipientAddress)

		transaction.SetVersion(umi.TxV6ActivateTransit)
		transaction.SetRecipient(recipient)

	case umi.TxDeactivateTransit:
		recipient, _ := umi.ParseAddress(*request.RecipientAddress)

		transaction.SetVersion(umi.TxV7DeactivateTransit)
		transaction.SetRecipient(recipient)
	}

	copy(transaction[85:149], ed25519.Sign(*request.Seed, transaction[0:85]))

	return transaction
}
