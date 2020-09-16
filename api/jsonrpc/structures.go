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

package jsonrpc

import (
	"encoding/binary"
	"encoding/json"

	"github.com/umitop/libumi"
)

type iStructure interface {
	GetStructureByPrefix(pfx []byte) (raw []byte, err error)
	ListStructures() (raws [][]byte, err error)
}

// ListStructures ...
func ListStructures(bc iBlockchain, _ []byte) (result []byte, errors []byte) {
	raws, err := bc.ListStructures()
	if err != nil {
		return nil, errServiceUnavail
	}

	return marshalStructures(raws), nil
}

// GetStructure ...
func GetStructure(bc iBlockchain, params []byte) (result []byte, errors []byte) {
	prm := new(struct {
		Prefix string `json:"prefix"`
	})

	if err := json.Unmarshal(params, prm); err != nil {
		return nil, errInvalidParams
	}

	if len(prm.Prefix) != 3 { //nolint:gomnd
		return nil, errInvalidParams
	}

	raw, err := bc.GetStructureByPrefix([]byte(prm.Prefix))
	if err != nil {
		return nil, errServiceUnavail
	}

	return marshalStructure(raw), nil
}

func marshalStructures(raws [][]byte) []byte {
	arr := make([]json.RawMessage, len(raws))

	for i, raw := range raws {
		arr[i] = marshalStructure(raw)
	}

	jsn, _ := json.Marshal(arr)

	return jsn
}

func marshalStructure(raw structure) []byte {
	jsn, _ := json.Marshal(struct {
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
		AddressCount     uint64   `json:"address_count"`
	}{
		Prefix:           raw.prefix(),
		Name:             raw.name(),
		FeePercent:       raw.feePercent(),
		ProfitPercent:    raw.profitPercent(),
		DepositPercent:   raw.depositPercent(),
		FeeAddress:       raw.feeAddress(),
		ProfitAddress:    raw.profitAddress(),
		MasterAddress:    raw.masterAddress(),
		TransitAddresses: raw.transitAddresses(),
		Balance:          raw.balance(),
		AddressCount:     raw.addressCount(),
	})

	return jsn
}

type structure []byte

func (s structure) prefix() string {
	return string(s[0:3])
}

func (s structure) name() string {
	return string(s[4 : 4+s[3]])
}

func (s structure) feePercent() uint16 {
	return binary.BigEndian.Uint16(s[38:40])
}

func (s structure) profitPercent() uint16 {
	return binary.BigEndian.Uint16(s[40:42])
}

func (s structure) depositPercent() uint16 {
	return binary.BigEndian.Uint16(s[42:44])
}

func (s structure) feeAddress() string {
	return (libumi.Address)(s[44:78]).Bech32()
}

func (s structure) profitAddress() string {
	return (libumi.Address)(s[78:112]).Bech32()
}

func (s structure) masterAddress() string {
	return (libumi.Address)(s[112:146]).Bech32()
}

func (s structure) transitAddresses() (t []string) {
	l := int(binary.BigEndian.Uint64(s[162:170]))
	if l == 0 {
		return t
	}

	t = make([]string, l)

	for i := 0; i < l; i++ {
		a := 170 + (34 * i) //nolint:gomnd
		t[i] = (libumi.Address)(s[a : a+34]).Bech32()
	}

	return t
}

func (s structure) balance() uint64 {
	return binary.BigEndian.Uint64(s[146:154])
}

func (s structure) addressCount() uint64 {
	return binary.BigEndian.Uint64(s[154:162])
}
