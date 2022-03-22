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
	"sync"

	"gitlab.com/umitop/umid/pkg/config"
	"gitlab.com/umitop/umid/pkg/umi"
)

const (
	firstJun2020 = 1590969600 // Monday, 1 June 2020 00:00:00
)

type Ledger struct {
	sync.RWMutex

	config       *config.Config
	accounts     map[umi.Prefix]map[umi.Address]*Account
	structures   map[umi.Prefix]*Structure
	transactions map[umi.Hash]struct{}
	nfts         map[umi.Hash]umi.Address

	LastBlockTimestamp    uint32
	LastBlockHeight       uint32
	LastBlockHash         umi.Hash
	LastTransactionHeight uint64
}

func NewLedger(conf *config.Config) *Ledger {
	ledger := &Ledger{
		config:       conf,
		accounts:     make(map[umi.Prefix]map[umi.Address]*Account),
		structures:   make(map[umi.Prefix]*Structure),
		transactions: make(map[umi.Hash]struct{}),
		nfts:         make(map[umi.Hash]umi.Address),
	}

	// Структуру UMI существует по умолчанию
	ledger.structures[umi.PfxVerUmi] = &Structure{
		accountType: umi.Umi,
		Prefix:      umi.PfxVerUmi,
		Description: "UMI",
		CreatedAt:   firstJun2020,
	}

	return ledger
}

func (ledger *Ledger) Account(address umi.Address) (account *Account, ok bool) {
	ledger.RLock()
	defer ledger.RUnlock()

	if accounts, ok1 := ledger.accounts[address.Prefix()]; ok1 {
		account, ok = accounts[address]
	}

	return account, ok
}

/*
func (ledger *Ledger) Balance(address umi.Address) (balance uint64) {
	if account, ok := ledger.Account(address); ok {
		timestamp := uint32(time.Now().Unix())
		balance = account.BalanceAt(timestamp)
	}

	return balance
}
*/

func (ledger *Ledger) Structures() (structures []*Structure) {
	ledger.RLock()
	defer ledger.RUnlock()

	structures = make([]*Structure, 0, len(ledger.structures))

	for _, structure := range ledger.structures {
		structures = append(structures, structure)
	}

	return structures
}

func (ledger *Ledger) Structure(prefix umi.Prefix) (structure *Structure, ok bool) {
	ledger.RLock()
	defer ledger.RUnlock()

	structure, ok = ledger.structures[prefix]

	return structure, ok
}

func (ledger *Ledger) HasTransaction(hash umi.Hash) bool {
	ledger.RLock()
	defer ledger.RUnlock()

	_, ok := ledger.transactions[hash]

	return ok
}

func (ledger *Ledger) NftsByAddr(addr umi.Address) []umi.Hash {
	ledger.RLock()
	defer ledger.RUnlock()

	nfts := make([]umi.Hash, 0)

	for hsh, adr := range ledger.nfts {
		if adr == addr {
			nfts = append(nfts, hsh)
		}
	}

	return nfts
}
