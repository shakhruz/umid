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
	"umid/umid"
)

func (s *postgres) AddTransaction(b []byte) error {
	_, err := s.conn.Exec(context.Background(), `select add_transaction($1)`, b)

	return err
}

func (s *postgres) TransactionsByAddress(adr []byte) (txs []*umid.Transaction2, err error) {
	rows, err := s.conn.Query(context.Background(), `select * from get_address_transactions($1, $2)`, adr, 100)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*umid.Transaction2, 0, 100)

	for rows.Next() {
		tx := &umid.Transaction2{}

		err := rows.Scan(
			&tx.Hash, &tx.Height, &tx.ConfirmedAt, &tx.BlockHeight, &tx.BlockTxIdx, &tx.Version, &tx.Sender,
			&tx.Recipient, &tx.Value, &tx.FeeAddress, &tx.FeeValue, &tx.Structure,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, tx)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return res, nil
}
