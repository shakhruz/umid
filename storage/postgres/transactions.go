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
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/umitop/libumi"
)

type strz struct {
	Prefix        string `json:"prefix"`
	Name          string `json:"name"`
	ProfitPercent uint16 `json:"profit_percent"`
	FeePercent    uint16 `json:"fee_percent"`
}

// AddTxToMempool ...
func (s *Postgres) AddTxToMempool(b []byte) error {
	_, err := s.conn.Exec(context.Background(), `select add_transaction($1)`, b)

	return err
}

// ListTxsByAddressBeforeKey ...
func (s *Postgres) ListTxsByAddressBeforeKey(adr []byte, _ []byte, lim uint16) (raws [][]byte, err error) {
	rows, err := s.conn.Query(context.Background(), `select * from get_address_transactions($1, $2)`, adr, lim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	raws = make([][]byte, 0, 100)

	for rows.Next() {
		raw, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}

		raws = append(raws, raw)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return raws, nil
}

// ListTxsByAddressAfterKey ...
func (s *Postgres) ListTxsByAddressAfterKey(_ []byte, _ []byte, _ uint16) ([][]byte, error) {
	return nil, nil
}

func scanTransaction(row pgx.Row) ([]byte, error) {
	var (
		hsh, snd, rcp, feeAdr []byte
		hgt, blkHgt           uint64
		tim                   time.Time
		txIdx                 uint16
		ver                   uint8
		val, feeVal           *uint64
		str                   strz
	)

	if err := row.Scan(&hsh, &hgt, &tim, &blkHgt, &txIdx, &ver, &snd, &rcp, &val, &feeAdr, &feeVal, &str); err != nil {
		return nil, err
	}

	raw := newTx()
	raw.setHash(hsh)
	raw.setTxHeight(hgt)
	raw.setVersion(ver)
	raw.setSender(snd)
	raw.setRecipient(rcp)
	raw.setValue(val)
	raw.setConfirmedAt(tim)
	raw.setBlockHeight(blkHgt)
	raw.setTxIndex(txIdx)
	raw.setFeeValue(feeVal)
	raw.setFeeAddress(feeAdr)

	switch ver {
	case libumi.CreateStructure, libumi.UpdateStructure:
		raw.setStructure(str)
	}

	return raw, nil
}

type tx []byte

func newTx() tx {
	return make(tx, 246)
}

func (t tx) setHash(h []byte) {
	copy(t[206:238], h)
}

func (t tx) setVersion(v uint8) {
	t[0] = v
}

func (t tx) setSender(b []byte) {
	(libumi.TransactionBuilder)(t).SetSender(b)
}

func (t tx) setRecipient(b []byte) {
	if len(b) != 0 {
		(libumi.TransactionBuilder)(t).SetRecipient(b)
	}
}

func (t tx) setValue(n *uint64) {
	if n != nil {
		(libumi.TransactionBuilder)(t).SetValue(*n)
	}
}

func (t tx) setConfirmedAt(tm time.Time) {
	binary.BigEndian.PutUint32(t[160:164], uint32(tm.Unix()))
}

func (t tx) setBlockHeight(n uint64) {
	binary.BigEndian.PutUint64(t[150:158], n)
}

func (t tx) setTxIndex(n uint16) {
	binary.BigEndian.PutUint16(t[158:160], n)
}

func (t tx) setFeeValue(n *uint64) {
	if n != nil {
		binary.BigEndian.PutUint64(t[164:172], *n)
	}
}

func (t tx) setFeeAddress(b []byte) {
	copy(t[172:206], b)
}

func (t tx) setTxHeight(n uint64) {
	binary.BigEndian.PutUint64(t[238:246], n)
}

func (t tx) setStructure(str strz) {
	(libumi.TransactionBuilder)(t).SetPrefix(str.Prefix)
	(libumi.TransactionBuilder)(t).SetName(str.Name)
	(libumi.TransactionBuilder)(t).SetFeePercent(str.FeePercent)
	(libumi.TransactionBuilder)(t).SetProfitPercent(str.ProfitPercent)
}
