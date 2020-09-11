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

package umid

import (
	"context"
	"sync"
	"time"
)

// IStorage ...
type IStorage interface {
	Worker(context.Context, *sync.WaitGroup)
	Mempool() (IMempool, error)
	Balance([]byte) (*Balance, error)
	StructureByPrefix(string) (*Structure2, error)
	Structures() ([]*Structure2, error)
	TransactionsByAddress([]byte) ([]*Transaction2, error)
	LastBlockHeight() (uint32, error)
	AddBlock([]byte) error
	AddTransaction([]byte) error
	BlocksByHeight(uint64) ([][]byte, error)
}

// IBlockchain ...
type IBlockchain interface {
	Balance(string) (*Balance, error)
	AddTransaction([]byte) error
	AddBlock([]byte) error
	StructureByPrefix(string) (*Structure, error)
	Structures() ([]*Structure, error)
	TransactionsByAddress(string) ([]*Transaction, error)
	LastBlockHeight() (uint32, error)
	BlocksByHeight(uint64) ([][]byte, error)
	Mempool() (IMempool, error)
}

// IMempool ...
type IMempool interface {
	Next() bool
	Value() []byte
	Close()
}

// TxStruct ...
type TxStruct struct {
	Prefix *string `json:"prefix,omitempty"`
}

// Transaction ...
type Transaction struct {
	Hash        string    `json:"hash"`
	ConfirmedAt int64     `json:"confirmed_at,omitempty"`
	Height      int32     `json:"height,omitempty"`
	BlockHeight int32     `json:"block_height"`
	BlockTxIdx  int32     `json:"block_tx_idx"`
	Version     int16     `json:"version"`
	Sender      string    `json:"sender"`
	Recipient   string    `json:"recipient,omitempty"`
	Value       *int64    `json:"value,omitempty"`
	FeeAddress  string    `json:"fee_address,omitempty"`
	FeeValue    *int64    `json:"fee_value,omitempty"`
	Structure   *TxStruct `json:"structure,omitempty"`
}

// Block ...
type Block struct {
	Hash []byte
}

// Balance ...
type Balance struct {
	Confirmed   uint64  `json:"confirmed"`
	Interest    uint16  `json:"interest"`
	Unconfirmed uint64  `json:"unconfirmed"`
	Composite   *uint64 `json:"composite,omitempty"`
	Type        string  `json:"type"`
}

// Structure ...
type Structure struct {
	Prefix           string   `json:"prefix"`
	Name             string   `json:"name"`
	FeePercent       uint16   `json:"fee_percent"`
	ProfitPercent    uint16   `json:"profit_percent"`
	DepositPercent   uint16   `json:"deposit_percent"`
	FeeAddress       string   `json:"fee_address"`
	ProfitAddress    string   `json:"profit_address"`
	MasterAddress    string   `json:"master_address"`
	TransitAddresses []string `json:"transit_addresses,omitempty"`
	Balance          uint64   `json:"balance"`
	AddressCount     uint32   `json:"address_count"`
}

// Transaction2 ...
type Transaction2 struct {
	Hash        []byte
	ConfirmedAt time.Time
	Height      int32
	BlockHeight int32
	BlockTxIdx  int32
	Version     int16
	Sender      []byte
	Recipient   []byte
	Value       *int64
	FeeAddress  []byte
	FeeValue    *int64
	Structure   *TxStruct
}

// Structure2 ...
type Structure2 struct {
	Prefix           string   `json:"prefix"`
	Name             string   `json:"name"`
	FeePercent       uint16   `json:"fee_percent"`
	ProfitPercent    uint16   `json:"profit_percent"`
	DepositPercent   uint16   `json:"deposit_percent"`
	FeeAddress       []byte   `json:"fee_address"`
	ProfitAddress    []byte   `json:"profit_address"`
	MasterAddress    []byte   `json:"master_address"`
	TransitAddresses [][]byte `json:"transit_addresses,omitempty"`
	Balance          uint64   `json:"balance"`
	AddressCount     uint32   `json:"address_count"`
}
