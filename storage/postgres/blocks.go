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
	"log"
)

func (s *postgres) LastBlockHeight() (n uint32, err error) {
	row := s.conn.QueryRow(context.Background(), `select coalesce(max(height), 0) from block`)
	err = row.Scan(&n)

	return
}

func (s *postgres) AddBlock(b []byte) error {
	var n int64
	err := s.conn.QueryRow(context.Background(), `select coalesce(add_block($1), 0)`, b).Scan(&n)

	if n != 0 {
		if n%1000 == 0 {
			log.Printf(`block %d added`, n)
		}
	}

	return err
}

func (s *postgres) BlocksByHeight(n uint64) ([][]byte, error) {
	const sql = `select lo_get(height) from block where height >= $1 and confirmed is true order by height limit 5000`

	rows, err := s.conn.Query(context.Background(), sql, n)
	if err != nil {
		return nil, err
	}

	res := make([][]byte, 0, 5000)

	for rows.Next() {
		b := make([]byte, 0)

		err := rows.Scan(&b)
		if err != nil {
			return nil, err
		}

		res = append(res, b)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return res, nil
}
