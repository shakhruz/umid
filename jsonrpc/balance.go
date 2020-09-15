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
	"log"

	"github.com/umitop/libumi"
)

type iBalance interface {
	GetBalance(adr []byte) (raw []byte, err error)
}

// GetBalance ...
func GetBalance(bc iBlockchain, params []byte) (result []byte, errors []byte) {
	prm := new(struct {
		Address string `json:"address"`
	})

	if err := json.Unmarshal(params, prm); err != nil {
		return nil, errInvalidParams
	}

	adr, err := libumi.NewAddressFromBech32(prm.Address)
	if err != nil {
		return nil, marshalError(codeInvalidParams, err.Error())
	}

	raw, err := bc.GetBalance(adr)
	if err != nil {
		log.Print(err.Error())

		return nil, errServiceUnavail
	}

	return marshalBalance(raw), nil
}

func marshalBalance(b balance) []byte {
	jsn, _ := json.Marshal(struct {
		Confirmed   uint64  `json:"confirmed"`
		Interest    uint16  `json:"interest"`
		Unconfirmed uint64  `json:"unconfirmed"`
		Composite   *uint64 `json:"composite,omitempty"`
		AddrType    string  `json:"address_type"`
	}{
		Confirmed:   b.confirmed(),
		Interest:    b.interest(),
		Unconfirmed: b.unconfirmed(),
		Composite:   b.composite(),
		AddrType:    b.addressType(),
	})

	return jsn
}

type balance []byte

func (b balance) confirmed() uint64 {
	return binary.BigEndian.Uint64(b[0:8])
}

func (b balance) interest() uint16 {
	return binary.BigEndian.Uint16(b[8:10])
}

func (b balance) unconfirmed() uint64 {
	return binary.BigEndian.Uint64(b[10:18])
}

func (b balance) composite() *uint64 {
	switch b.addressType() {
	case "dev", "profit":
		n := binary.BigEndian.Uint64(b[18:26])

		return &n
	}

	return nil
}

func (b balance) addressType() string {
	t := map[uint8]string{
		0: "dev",
		1: "master",
		2: "profit",
		3: "fee",
		4: "transit",
		5: "deposit",
		6: "umi",
	}

	return t[b[26]]
}
