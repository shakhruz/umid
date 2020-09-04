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

package blockchain

import (
	"umid/umid"

	"github.com/umitop/libumi"
)

// StructureByPrefix ...
func (bc *Blockchain) StructureByPrefix(p string) (*umid.Structure, error) {
	s, err := bc.storage.StructureByPrefix(p)
	if err != nil {
		return nil, err
	}

	return structConvert(s), nil
}

// Structures ...
func (bc *Blockchain) Structures() ([]*umid.Structure, error) {
	s, err := bc.storage.Structures()
	if err != nil {
		return nil, err
	}

	res := make([]*umid.Structure, len(s))

	for i, raw := range s {
		res[i] = structConvert(raw)
	}

	return res, nil
}

func structConvert(s *umid.Structure2) *umid.Structure {
	res := &umid.Structure{
		Prefix:         s.Prefix,
		Name:           s.Name,
		FeePercent:     s.FeePercent,
		ProfitPercent:  s.ProfitPercent,
		DepositPercent: s.DepositPercent,
		FeeAddress:     (*libumi.Address)(&s.FeeAddress).Bech32(),
		ProfitAddress:  (*libumi.Address)(&s.ProfitAddress).Bech32(),
		MasterAddress:  (*libumi.Address)(&s.MasterAddress).Bech32(),
		Balance:        s.Balance,
		AddressCount:   s.AddressCount,
	}

	if s.TransitAddresses != nil && len(s.TransitAddresses) > 0 {
		res.TransitAddresses = make([]string, len(s.TransitAddresses))

		for i := range s.TransitAddresses {
			res.TransitAddresses[i] = (*libumi.Address)(&s.TransitAddresses[i]).Bech32()
		}
	}

	return res
}
