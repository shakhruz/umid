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
	"log"

	"github.com/jackc/pgx/v4"
)

// ListStructures ...
func (s *Postgres) ListStructures() (raws [][]byte, err error) {
	rows, err := s.conn.Query(context.Background(), `select * from get_structures()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	raws = make([][]byte, 0)

	for rows.Next() {
		raw, err := scanStructure(rows)
		if err != nil {
			return nil, err
		}

		raws = append(raws, raw)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())

		return nil, err
	}

	return raws, nil
}

// GetStructureByPrefix ...
func (s *Postgres) GetStructureByPrefix(pfx []byte) (raw []byte, err error) {
	row := s.conn.QueryRow(context.Background(), `select * from get_structures_by_prefix($1)`, pfx)

	return scanStructure(row)
}

func scanStructure(row pgx.Row) ([]byte, error) {
	var (
		pfx, name              string
		bal, adrCnt            uint64
		feePer, prfPer, depPer uint16
		feeAdr, prfAdr, mstAdr []byte
		trnAdrs                [][]byte
	)

	err := row.Scan(&pfx, &name, &feePer, &prfPer, &depPer, &feeAdr, &prfAdr, &mstAdr, &trnAdrs, &bal, &adrCnt)
	if err != nil {
		return nil, err
	}

	raw := newStructure()
	raw.setPrefix(pfx)
	raw.setName(name)
	raw.setFeePercent(feePer)
	raw.setProfitPercent(prfPer)
	raw.setDepositPercent(depPer)
	raw.setFeeAddress(feeAdr)
	raw.setProfitAddress(prfAdr)
	raw.setMasterAddress(mstAdr)
	raw.setTransitAddresses(trnAdrs)
	raw.setBalance(bal)
	raw.setAddressCount(adrCnt)

	return raw, nil
}

type structure []byte

func newStructure() structure {
	return make(structure, 170)
}

func (s structure) setPrefix(pfx string) {
	copy(s[0:3], pfx)
}

func (s structure) setName(name string) {
	s[3] = uint8(len(name))
	copy(s[4:38], name)
}

func (s structure) setFeePercent(n uint16) {
	binary.BigEndian.PutUint16(s[38:40], n)
}

func (s structure) setProfitPercent(n uint16) {
	binary.BigEndian.PutUint16(s[40:42], n)
}

func (s structure) setDepositPercent(n uint16) {
	binary.BigEndian.PutUint16(s[42:44], n)
}

func (s structure) setFeeAddress(b []byte) {
	copy(s[44:78], b)
}

func (s structure) setProfitAddress(b []byte) {
	copy(s[78:112], b)
}

func (s structure) setMasterAddress(b []byte) {
	copy(s[112:146], b)
}

func (s *structure) setTransitAddresses(b [][]byte) {
	st := *s

	binary.BigEndian.PutUint64(st[162:170], uint64(len(b)))

	for _, x := range b {
		st = append(st, x...)
	}

	*s = st
}

func (s structure) setBalance(n uint64) {
	binary.BigEndian.PutUint64(s[146:154], n)
}

func (s structure) setAddressCount(n uint64) {
	binary.BigEndian.PutUint64(s[154:162], n)
}
