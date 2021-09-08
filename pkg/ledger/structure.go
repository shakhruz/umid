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
	"encoding/json"
	"math"
	"time"

	"gitlab.com/umitop/umid/pkg/openlibm"
	"gitlab.com/umitop/umid/pkg/umi"
)

type Structure struct {
	accountType umi.AccountType
	CreatedAt   uint32

	Prefix        umi.Prefix
	Description   string
	ProfitPercent uint16
	FeePercent    uint16

	MasterAddress umi.Address
	FeeAddress    umi.Address
	ProfitAddress umi.Address
	DevAddress    umi.Address

	AddressCount int

	Balance   uint64
	UpdatedAt uint32

	Level             uint8
	LevelInterestRate uint16
}

func NewStructure(network string, prefix umi.Prefix, masterAddr umi.Address) *Structure {
	return &Structure{
		accountType:   umi.Deposit,
		Prefix:        prefix,
		MasterAddress: masterAddr,
		FeeAddress:    newStructureAddr(prefix, masterAddr),
		ProfitAddress: newStructureAddr(prefix, masterAddr),
		DevAddress:    newDevAddr(network, prefix),
	}
}

func (structure *Structure) IsOwner(addr umi.Address) bool {
	return structure.MasterAddress == addr
}

func (structure *Structure) InterestRate(accountType umi.AccountType) uint16 {
	if structure.LevelInterestRate == 0 {
		return 0
	}

	switch accountType {
	case umi.Genesis, umi.Umi:
		return 0
	case umi.Deposit, umi.Transit:
		return structure.LevelInterestRate - structure.ProfitPercent
	case umi.Fee, umi.Profit:
		return structure.LevelInterestRate
	case umi.Dev:
		return structure.LevelInterestRate + 2_00
	default:
		return 0
	}
}

func (structure *Structure) IncreaseBalance(amount uint64, timestamp uint32) {
	structure.updateBalance(timestamp)
	structure.Balance += amount
}

func (structure *Structure) DecreaseBalance(amount uint64, timestamp uint32) {
	structure.updateBalance(timestamp)

	if structure.Balance < amount {
		structure.Balance = amount
	}

	structure.Balance -= amount
}

func (structure *Structure) BalanceAt(timestamp uint32) uint64 {
	if structure.LevelInterestRate == 0 {
		return structure.Balance
	}

	r := float64(1) + (float64(structure.LevelInterestRate-structure.ProfitPercent) / float64(100_00))
	n := float64(1) / float64(2592000)
	growthRate := math.Pow(r, n)
	balance := math.Floor(float64(structure.Balance) * openlibm.Pow(growthRate, float64(timestamp-structure.UpdatedAt)))

	return uint64(balance)
}

func (structure *Structure) updateBalance(timestamp uint32) {
	if timestamp == structure.UpdatedAt {
		return
	}

	structure.Balance = structure.BalanceAt(timestamp)
	structure.UpdatedAt = timestamp
}

func (structure *Structure) MarshalJSON() ([]byte, error) {
	data := struct {
		Prefix       string `json:"prefix"`
		Description  string `json:"description"`
		Balance      uint64 `json:"balance"`
		CreatedAt    string `json:"createdAt"`
		UpdatedAt    string `json:"updatedAt"`
		AddressCount int    `json:"addressCount"`

		ProfitPercent *uint16 `json:"profitPercent,omitempty"`
		FeePercent    *uint16 `json:"feePercent,omitempty"`
		MasterAddress *string `json:"masterAddress,omitempty"`
		DevAddress    *string `json:"devAddress,omitempty"`
		FeeAddress    *string `json:"feeAddress,omitempty"`
		ProfitAddress *string `json:"profitAddress,omitempty"`
		Level         *uint8  `json:"level,omitempty"`
		InterestRate  *uint16 `json:"interestRate,omitempty"`
	}{
		Prefix:       structure.Prefix.String(),
		Description:  structure.Description,
		Balance:      structure.BalanceAt(uint32(time.Now().Unix())),
		AddressCount: structure.AddressCount,
		CreatedAt:    time.Unix(int64(structure.CreatedAt), 0).UTC().Format(time.RFC3339),
		UpdatedAt:    time.Unix(int64(structure.UpdatedAt), 0).UTC().Format(time.RFC3339),
	}

	if structure.Prefix != umi.PfxVerUmi {
		data.ProfitPercent = new(uint16)
		*data.ProfitPercent = structure.ProfitPercent

		data.FeePercent = new(uint16)
		*data.FeePercent = structure.FeePercent

		data.MasterAddress = new(string)
		*data.MasterAddress = structure.MasterAddress.String()

		data.DevAddress = new(string)
		*data.DevAddress = structure.DevAddress.String()

		data.FeeAddress = new(string)
		*data.FeeAddress = structure.FeeAddress.String()

		data.ProfitAddress = new(string)
		*data.ProfitAddress = structure.ProfitAddress.String()

		data.Level = new(uint8)
		*data.Level = structure.Level

		data.InterestRate = new(uint16)
		*data.InterestRate = structure.LevelInterestRate
	}

	return json.Marshal(data) //nolint:wrapcheck // ...
}

func newStructureAddr(prefix umi.Prefix, addr umi.Address) umi.Address {
	var address umi.Address

	address.SetPrefix(prefix)
	address.SetPublicKey(addr.PublicKey())

	return address
}

func newDevAddr(network string, prefix umi.Prefix) umi.Address {
	// umi16uz7khspwq0patw777wgn7hgk6pvds2sxqgwt546z5n489mwmj2szdn2h5
	publicKey := []byte{
		0xd7, 0x05, 0xeb, 0x5e, 0x01, 0x70, 0x1e, 0x1e, 0xad, 0xde, 0xf7, 0x9c, 0x89, 0xfa, 0xe8, 0xb6,
		0x82, 0xc6, 0xc1, 0x50, 0x30, 0x10, 0xe5, 0xd2, 0xba, 0x15, 0x27, 0x53, 0x97, 0x6e, 0xdc, 0x95,
	}

	if network == "testnet" {
		// umi16dhtrj348vaa63lp46u24hs5mjjjxzwqn75qwvnzke6uyr5txukqgckvra
		publicKey = []byte{
			0xd3, 0x6e, 0xb1, 0xca, 0x35, 0x3b, 0x3b, 0xdd, 0x47, 0xe1, 0xae, 0xb8, 0xaa, 0xde, 0x14, 0xdc,
			0xa5, 0x23, 0x09, 0xc0, 0x9f, 0xa8, 0x07, 0x32, 0x62, 0xb6, 0x75, 0xc2, 0x0e, 0x8b, 0x37, 0x2c,
		}
	}

	var address umi.Address

	address.SetPrefix(prefix)
	address.SetPublicKey(publicKey)

	return address
}
