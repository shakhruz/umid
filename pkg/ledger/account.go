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
	"log"
	"math"

	"gitlab.com/umitop/umid/pkg/openlibm"
	"gitlab.com/umitop/umid/pkg/umi"
)

type Account struct {
	Type         umi.AccountType
	Balance      uint64
	UpdatedAt    uint32
	InterestRate uint16

	TransactionCount uint64

	growthRate float64
}

func NewAccount(structure *Structure) *Account {
	account := &Account{
		Type: structure.accountType,
	}

	account.SetInterestRate(structure.InterestRate(account.Type), 0)

	return account
}

func (account *Account) BalanceAt(timestamp uint32) uint64 {
	if account.InterestRate == 0 || account.UpdatedAt > timestamp {
		return account.Balance
	}

	balance := math.Floor(float64(account.Balance) *
		openlibm.Pow(account.growthRate, float64(timestamp-account.UpdatedAt)))

	return uint64(balance)
}

func (account *Account) IncreaseBalance(amount uint64, timestamp uint32) {
	account.UpdateBalance(timestamp)
	account.Balance += amount
}

func (account *Account) DecreaseBalance(amount uint64, timestamp uint32) bool {
	account.UpdateBalance(timestamp)

	if account.Balance < amount {
		log.Printf("недостаточно монеток %s %d %d", account.Type.String(), account.Balance, amount)

		return false
	}

	account.Balance -= amount

	return true
}

func (account *Account) SetType(accountType umi.AccountType) {
	account.Type = accountType
}

func (account *Account) SetInterestRate(interest uint16, timestamp uint32) {
	account.UpdateBalance(timestamp)
	account.InterestRate = interest

	r := float64(1) + (float64(interest) / float64(100_00))
	n := float64(1) / float64(2592000)

	account.growthRate = math.Pow(r, n)
}

func (account *Account) UpdateBalance(timestamp uint32) {
	if account.UpdatedAt == timestamp {
		return
	}

	account.Balance = account.BalanceAt(timestamp)
	account.UpdatedAt = timestamp
}
