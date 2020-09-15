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

package postgres

import (
	"context"
	"encoding/binary"
)

// GetBalance ...
func (s *Postgres) GetBalance(adr []byte) (raw []byte, err error) {
	var (
		confirmed   uint64
		interest    uint16
		unconfirmed uint64
		composite   *uint64
		addressType string
	)

	row := s.conn.QueryRow(context.Background(), `select * from get_address_balance($1)`, adr)

	if err := row.Scan(&confirmed, &interest, &unconfirmed, &composite, &addressType); err != nil {
		return nil, err
	}

	bal := newBalance()
	bal.setConfirmed(confirmed)
	bal.setInterest(interest)
	bal.setUnconfirmed(unconfirmed)
	bal.setComposite(composite)
	bal.setAddressType(addressType)

	return bal, nil
}

type balance []byte

func newBalance() balance {
	return make(balance, 27)
}

func (b balance) setConfirmed(n uint64) {
	binary.BigEndian.PutUint64(b[0:8], n)
}

func (b balance) setInterest(n uint16) {
	binary.BigEndian.PutUint16(b[8:10], n)
}

func (b balance) setUnconfirmed(n uint64) {
	binary.BigEndian.PutUint64(b[10:18], n)
}

func (b balance) setComposite(n *uint64) {
	if n != nil {
		binary.BigEndian.PutUint64(b[18:26], *n)
	}
}

func (b balance) setAddressType(t string) {
	//nolint:gomnd
	v := map[string]uint8{
		"dev":     0,
		"master":  1,
		"profit":  2,
		"fee":     3,
		"transit": 4,
		"deposit": 5,
		"umi":     6,
	}

	b[26] = v[t]
}
